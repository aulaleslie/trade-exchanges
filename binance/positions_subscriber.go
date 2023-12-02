package binance

import (
	"context"
	"encoding/json"
	"time"

	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/binance/adshao_binance"
	"github.com/cockroachdb/apd"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const accountUpdateEventType string = "outboundAccountPosition"
const accountUpdateFuturesEventType string = "ACCOUNT_UPDATE"

type AccountUpdate struct {
	EventType     string                    `json:"e"`
	EventTime     int64                     `json:"E"`
	BalancesArray []BinanceBalancePositions `json:"B"`
}

type BinanceBalancePositions struct {
	Asset  string `json:"a"`
	Free   string `json:"f"`
	Locked string `json:"l"`
}

type FuturesAccountUpdateData struct {
	EventReasonType string                    `json:"m"`
	Positions       []FuturesBinancePositions `json:"P"`
}

type FuturesAccountUpdate struct {
	EventType string                   `json:"e"`
	EventTime int64                    `json:"E"`
	Data      FuturesAccountUpdateData `json:"a"`
}

type FuturesBinancePositions struct {
	Symbol         string `json:"s"`
	PositionAmount string `json:"pa"`
}

func SubscribeToPositionsFutures(
	ctx context.Context,
	wsEndpoint string,
	lg *zap.Logger,
) (<-chan exchanges.PositionEvent, error) {
	lg.Info("calling SubscribeToPositions Futures ...")

	cfg := adshao_binance.WSConfig{
		Endpoint:  wsEndpoint,
		KeepAlive: true,
		Timeout:   30 * time.Second,
	}

	wsServeCtx, cancel := context.WithCancel(ctx)
	in, err := adshao_binance.WSServe(wsServeCtx, &cfg, lg.Named("Positions"))
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, "can't start websocket")
	}

	out := make(chan exchanges.PositionEvent, 100) // TODO: move to config
	go func() {
		defer cancel()
		defer close(out)

		for msg := range in {
			lg.Sugar().Infof("DisconnectedWithErr: %s", msg.DisconnectedWithErr)
			lg.Sugar().Infof("Payload: %s", string(msg.Payload))

			if msg.DisconnectedWithErr != nil {
				lg.Sugar().Errorf("error while reading event: %v", msg.DisconnectedWithErr)
				out <- exchanges.PositionEvent{
					DisconnectedWithErr: msg.DisconnectedWithErr,
				}
				return
			}

			ok, err := isAccountUpdateFuturesEventPayload(msg.Payload)
			if err != nil {
				lg.Sugar().Errorf("error while checking account update event type: %v", err)

				out <- exchanges.PositionEvent{
					DisconnectedWithErr: errors.Wrap(err, "can't understand type of account update datastream event"),
				}
				return
			}

			if !ok {
				continue
			}

			result, err := mapToAccountUpdateFuturesEventPayload(msg.Payload)
			if err != nil {
				out <- exchanges.PositionEvent{
					DisconnectedWithErr: errors.Wrap(err, "can't parse OrderEvent"),
				}
				return
			}

			out <- exchanges.PositionEvent{
				Payload: result,
			}
		}
	}()

	return out, nil
}

func SubscribeToPositions(
	ctx context.Context,
	wsEndpoint string,
	lg *zap.Logger,
) (<-chan exchanges.PositionEvent, error) {
	lg.Info("calling SubscribeToPositions ...")

	cfg := adshao_binance.WSConfig{
		Endpoint:  wsEndpoint,
		KeepAlive: true,
		Timeout:   30 * time.Second,
	}

	wsServeCtx, cancel := context.WithCancel(ctx)
	in, err := adshao_binance.WSServe(wsServeCtx, &cfg, lg.Named("Positions"))
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, "can't start websocket")
	}

	out := make(chan exchanges.PositionEvent, 100) // TODO: move to config
	go func() {
		defer cancel()
		defer close(out)

		lg.Info("start processing event ...")
		for msg := range in {
			if msg.DisconnectedWithErr != nil {
				lg.Sugar().Errorf("error while reading event: %v", msg.DisconnectedWithErr)

				out <- exchanges.PositionEvent{
					DisconnectedWithErr: msg.DisconnectedWithErr,
				}
				return
			}

			ok, err := isAccountUpdateEventPayload(msg.Payload)
			if err != nil {
				lg.Sugar().Errorf("error while checking account update event type: %v", err)

				out <- exchanges.PositionEvent{
					DisconnectedWithErr: errors.Wrap(err, "can't understand type of account update datastream event"),
				}
				return
			}

			if !ok {
				continue
			}

			result, err := mapToAccountUpdateEventPayload(msg.Payload)
			if err != nil {
				out <- exchanges.PositionEvent{
					DisconnectedWithErr: errors.Wrap(err, "can't parse OrderEvent"),
				}
				return
			}

			out <- exchanges.PositionEvent{
				Payload: result,
			}
		}
	}()

	return out, nil
}

func mapToAccountUpdateEventPayload(message []byte) (p []*exchanges.PositionPayload, e error) {
	accountUpdate := AccountUpdate{}
	err := json.Unmarshal(message, &accountUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "can't unmarshal JSON")
	}

	positionPayload := make([]*exchanges.PositionPayload, 0)
	for _, balance := range accountUpdate.BalancesArray {
		freeBalance, _, err := apd.NewFromString(balance.Free)
		if err != nil {
			return nil, errors.Wrap(err, "can't generate free balance")
		}

		positionPayload = append(positionPayload, &exchanges.PositionPayload{
			Symbol: balance.Asset,
			Value:  freeBalance,
		})
	}

	return positionPayload, nil
}

func isAccountUpdateEventPayload(message []byte) (bool, error) {
	data := userDataStreamCommonMessage{}
	err := json.Unmarshal(message, &data)
	if err != nil {
		return false, errors.Wrap(err, string(message))
	}
	return data.EventType == accountUpdateEventType, nil
}

func isAccountUpdateFuturesEventPayload(message []byte) (bool, error) {
	data := userDataStreamCommonMessage{}
	err := json.Unmarshal(message, &data)
	if err != nil {
		return false, errors.Wrap(err, string(message))
	}
	return data.EventType == accountUpdateFuturesEventType, nil
}

func mapToAccountUpdateFuturesEventPayload(message []byte) (p []*exchanges.PositionPayload, e error) {
	accountUpdate := FuturesAccountUpdate{}
	err := json.Unmarshal(message, &accountUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "can't unmarshal JSON")
	}

	positionPayload := make([]*exchanges.PositionPayload, 0)
	for _, balance := range accountUpdate.Data.Positions {
		freeBalance, _, err := apd.NewFromString(balance.PositionAmount)
		if err != nil {
			return nil, errors.Wrap(err, "can't generate free balance")
		}

		positionPayload = append(positionPayload, &exchanges.PositionPayload{
			Symbol: balance.Symbol,
			Value:  freeBalance,
		})
	}

	return positionPayload, nil
}
