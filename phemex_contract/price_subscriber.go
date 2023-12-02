package phemex_contract

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/Krisa/go-phemex"
	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/utils"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type PriceSubscriber struct{}

// No reconnection in case of error
// Returns control after connect
func SubscribeToPrice(ctx context.Context, symbol string, lg *zap.Logger) (<-chan exchanges.PriceEvent, error) {
	ctxScales, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	err := ScalesSubscriberInstance.CheckOrUpdate(ctxScales)
	if err != nil {
		return nil, errors.Wrap(err, "can't fetch scales")
	}

	// https://github.com/phemex/phemex-api-docs/blob/master/Public-Contract-API-en.md#subscribe-24-hours-ticker
	callID := 75
	cfg := utils.WSConfig{
		Endpoint: "wss://phemex.com/ws",
		InitialTextMessage: []byte(`{
			"id":     ` + strconv.Itoa(callID) + `,
			"method": "market24h.subscribe",
			"params": []
		}`),
		KeepAlive:         true,
		Timeout:           15 * time.Second,
		HeartbeatInterval: 5 * time.Second,
	}

	wsServeCtx, cancel := context.WithCancel(ctx)
	in, err := utils.WSConnectAndWatch(wsServeCtx, &cfg, lg.Named("Prices"))
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, "can't start websocket")
	}

	out := make(chan exchanges.PriceEvent, 100) // TODO: move to config
	go func() {
		defer cancel()
		defer close(out)

		streamingEnabled := false

		for msg := range in {
			if msg.DisconnectedWithErr != nil {
				out <- exchanges.PriceEvent{
					DisconnectedWithErr: msg.DisconnectedWithErr,
				}
				return
			}

			market24HData, phemexWSError, err := mapWSMarket24H(msg.Payload)
			if err != nil {
				out <- exchanges.PriceEvent{DisconnectedWithErr: err}
				return
			}

			if phemexWSError != nil {
				passed, err := checkPhemexWSCallResponse(callID, phemexWSError)
				if err != nil {
					out <- exchanges.PriceEvent{DisconnectedWithErr: err}
					return
				}
				if passed {
					streamingEnabled = true
				}
				continue
			}

			if !streamingEnabled {
				continue
			}

			symbolScales, err := ScalesSubscriberInstance.GetLastSymbolScales(symbol)
			if err != nil {
				out <- exchanges.PriceEvent{DisconnectedWithErr: err}
				return
			}

			if market24HData.Market24H.Symbol == symbol {
				out <- exchanges.PriceEvent{
					Payload: utils.Div(market24HData.Market24H.CloseEp.Value, symbolScales.PriceScaleDivider)}
			}
		}
	}()

	return out, nil
}

func checkPhemexWSCallResponse(callerID int, phemexWSError *phemex.WsError) (passed bool, e error) {
	if phemexWSError.Error != nil {
		status := "-"
		if phemexWSError.Result != nil {
			status = phemexWSError.Result.Status
		}
		return false, errors.Errorf("came phemex WS error: %s (code=%d, id=%d, status=%s)",
			phemexWSError.Error.Message, phemexWSError.Error.Code, phemexWSError.ID, status)
	}

	return phemexWSError.ID == callerID, nil
}

func mapWSMarket24H(message []byte) (*WSMarket24Msg, *phemex.WsError, error) {
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

	var market24Msg *WSMarket24Msg
	err := json.Unmarshal(message, &market24Msg)
	if err != nil {
		return nil, nil, errors.Wrap(err, "can't unmarshall market 24 response")
	}
	if market24Msg.Market24H == nil {
		return nil, nil, errors.New("came response with empty 'market24h' field")
	}
	return market24Msg, nil, nil
}

type WSMarket24Msg struct {
	Market24H *WSMarket24Data `json:"market24h"`
	Timestamp int64           `json:"timestamp"`
}

type WSMarket24Data struct {
	CloseEp utils.APDJSON `json:"close"`  // : <close priceEp>; "close":  87425000,
	Symbol  string        `json:"symbol"` // : "<symbol>";      "symbol": "BTCUSD",

	// FundingRate     Type    `json:"fundingRate"`     // : <funding rateEr>; "fundingRate":  10000,
	// High            Type    `json:"high"`            // : <high priceEp>;   "high":         92080000,
	// IndexPrice      Type    `json:"indexPrice"`      // : <index priceEp>;  "indexPrice":   87450676,
	// Low             Type    `json:"low"`             // : <low priceEp>;    "low":          87130000,
	// MarkPrice       Type    `json:"markPrice"`       // : <mark priceEp>;   "markPrice":    87453092,
	// Open            Type    `json:"open"`            // : <open priceEp>;   "open":         90710000,
	// OpenInterest    Type    `json:"openInterest"`    // : <open interest>;  "openInterest": 7821141,
	// PredFundingRate Type    `json:"predFundingRate"` // : <predicated funding rateEr>; "predFundingRate": 7609,
	// Turnover        Type    `json:"turnover"`        // : <turnoverEv>";    "turnover":     1399362834123,
	// Volume          Type    `json:"volume"`          // : <volume>";        "volume":       125287131
}
