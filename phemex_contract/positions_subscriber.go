package phemex_contract

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Krisa/go-phemex"
	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/phemex_contract/krisa_phemex_fork"
	"github.com/aulaleslie/trade-exchanges/utils"
	"github.com/cockroachdb/apd"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type WSPositionEvent struct {
	event  exchanges.PositionEvent
	fields *orderFields
}

func SubscribeToPositions(ctx context.Context, client *phemex.Client, lg *zap.Logger) (<-chan WSPositionEvent, error) {
	conn, err := client.NewWsAuthService().Do(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to auth")
	}

	callID := 232
	cfg := utils.WSConfig{
		Endpoint: "wss://phemex.com/ws",
		InitialTextMessage: []byte(`{
			"id":     ` + strconv.Itoa(callID) + `,
			"method": "aop.subscribe",
			"params": []
		}`),
		KeepAlive:         true,
		Timeout:           15 * time.Second,
		HeartbeatInterval: 5 * time.Second,
	}

	wsServeCtx, cancel := context.WithCancel(ctx)
	in, err := utils.WSWatch(wsServeCtx, conn, &cfg, lg.Named("Positions"))
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, "can't start websocket")
	}

	subscribeErr := make(chan error, 1)
	out := make(chan WSPositionEvent, 100) // TODO: move to config
	go func() {
		defer cancel()
		defer close(out)

		subscribed := false

		errHandler := func(err error) {
			if !subscribed {
				subscribeErr <- err
				return
			}

			out <- WSPositionEvent{
				event: exchanges.PositionEvent{
					DisconnectedWithErr: err,
				},
			}
		}

		for msg := range in {
			if msg.DisconnectedWithErr != nil {
				errHandler(msg.DisconnectedWithErr)
				return
			}

			log.Printf("SubscribeToPositions payload: %v", string(msg.Payload))

			aop, phemexWSError, err := mapAOPPositionData(msg.Payload)
			if err != nil {
				errHandler(err)
				return
			}

			if phemexWSError != nil {
				passed, err := checkPhemexWSCallResponse(callID, phemexWSError)
				if err != nil {
					errHandler(err)
					return
				}
				if passed {
					subscribed = true
					close(subscribeErr)
				}
				continue
			}

			if !subscribed {
				continue
			}

			err = sendAOPPositionEvents(out, aop)
			if err != nil {
				errHandler(err)
				return
			}
		}
	}()

	return out, <-subscribeErr
}

func mapAOPPositionData(message []byte) (*phemex.WsAOP, *phemex.WsError, error) {
	if strings.Contains(string(message), `"error"`) {
		var callResponse *phemex.WsError
		err := json.Unmarshal(message, &callResponse)
		if err != nil {
			return nil, nil, errors.Wrap(err, "can't unmarshall call response")
		}
		if callResponse.Error != nil || callResponse.Result != nil {
			return nil, callResponse, nil
		}
	}

	var aop *phemex.WsAOP
	err := json.Unmarshal(message, &aop)
	if err != nil {
		return nil, nil, errors.Wrap(err, "can't unmarshall market 24 response")
	}
	return aop, nil, nil
}

// Don't write error to channel there
func sendAOPPositionEvents(ch chan<- WSPositionEvent, aop *phemex.WsAOP) error {
	for _, order := range aop.Orders {
		positionDatas := make([]*exchanges.PositionPayload, 0)
		for _, position := range aop.Positions {
			value := new(apd.Decimal)
			sizeFloat, err := value.SetFloat64(position.Size)
			if err != nil {
				return err
			}

			positionDatas = append(positionDatas, &exchanges.PositionPayload{
				Symbol: position.Symbol,
				Value:  sizeFloat,
			})
		}

		a := exchanges.PositionEvent{
			DisconnectedWithErr: nil,
			Reconnected:         nil,
			Payload:             positionDatas,
		}
		b := &orderFields{
			ClOrdID:     order.ClOrdID,
			OrdType:     krisa_phemex_fork.OrderType(order.OrdType),
			OrderQty:    order.OrderQty,
			PriceEp:     order.PriceEp,
			Side:        krisa_phemex_fork.SideType(order.Side),
			Symbol:      order.Symbol,
			TimeInForce: krisa_phemex_fork.TimeInForceType(order.TimeInForce),
		}
		ch <- WSPositionEvent{
			event:  a,
			fields: b,
		}
	}
	return nil
}
