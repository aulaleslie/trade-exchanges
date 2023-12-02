package binance

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/binance/adshao_binance"
	"github.com/aulaleslie/trade-exchanges/utils"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// SubscribeToPrice No reconnection in case of error
// Returns control after connect
func SubscribeToPrice(ctx context.Context, urls BinanceURLs, symbol string, lg *zap.Logger) (<-chan exchanges.PriceEvent, error) {
	symbol = strings.ToUpper(symbol)

	cfg := adshao_binance.WSConfig{
		Endpoint:  urls.WSAllMiniMarketsStatURL(), // OPTIMIZATION: use price data for single symbol, not for all
		KeepAlive: true,
		Timeout:   30 * time.Second,
	}

	wsServeCtx, cancel := context.WithCancel(ctx)
	in, err := adshao_binance.WSServe(wsServeCtx, &cfg, lg.Named("Prices"))
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, "can't start websocket")
	}

	out := make(chan exchanges.PriceEvent, 100) // TODO: move to config
	go func() {
		defer cancel()
		defer close(out)

		for msg := range in {
			if msg.DisconnectedWithErr != nil {
				out <- exchanges.PriceEvent{
					DisconnectedWithErr: msg.DisconnectedWithErr,
				}
				return
			}

			stats, err := mapWSAllMiniMarketsStatEvent(msg.Payload)
			if err != nil {
				out <- exchanges.PriceEvent{DisconnectedWithErr: err}
				return
			}

			for _, stat := range *stats {
				statSymbolNorm := strings.ToUpper(stat.Symbol)
				if statSymbolNorm == symbol {
					price, err := utils.FromStringErr(stat.LastPrice)
					if err != nil {
						out <- exchanges.PriceEvent{DisconnectedWithErr: err}
						return
					}
					out <- exchanges.PriceEvent{Payload: price}
					break
				}
			}
		}
	}()

	return out, nil
}

// Similar with SubscribeToPrice but it accepts ws endpoint instead of urls object
func SubscribeToPriceV2(ctx context.Context, wsEndpoint string, symbol string, lg *zap.Logger) (<-chan exchanges.PriceEvent, error) {
	symbol = strings.ToUpper(symbol)

	cfg := adshao_binance.WSConfig{
		Endpoint:  wsEndpoint, // OPTIMIZATION: use price data for single symbol, not for all
		KeepAlive: true,
		Timeout:   30 * time.Second,
	}

	wsServeCtx, cancel := context.WithCancel(ctx)
	in, err := adshao_binance.WSServe(wsServeCtx, &cfg, lg.Named("Prices"))
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, "can't start websocket")
	}

	out := make(chan exchanges.PriceEvent, 100) // TODO: move to config
	go func() {
		defer cancel()
		defer close(out)

		for msg := range in {
			if msg.DisconnectedWithErr != nil {
				out <- exchanges.PriceEvent{
					DisconnectedWithErr: msg.DisconnectedWithErr,
				}
				return
			}

			stats, err := mapWSAllMiniMarketsStatEvent(msg.Payload)
			if err != nil {
				out <- exchanges.PriceEvent{DisconnectedWithErr: err}
				return
			}

			for _, stat := range *stats {
				statSymbolNorm := strings.ToUpper(stat.Symbol)
				if statSymbolNorm == symbol {
					price, err := utils.FromStringErr(stat.LastPrice)
					if err != nil {
						out <- exchanges.PriceEvent{DisconnectedWithErr: err}
						return
					}
					out <- exchanges.PriceEvent{Payload: price}
					break
				}
			}
		}
	}()

	return out, nil
}

type WSAllMiniMarketsStatEvent []*WSMiniMarketsStatEvent

// WSMiniMarketsStatEvent define websocket market mini-ticker statistics event
type WSMiniMarketsStatEvent struct {
	// ! DON'T REMOVE NullJSONValue fields. They are used to make JSON case-senitive
	Event string `json:"e"`
	Time  int64  `json:"E"`

	Symbol  string              `json:"s"`
	Ignore1 utils.NullJSONValue `json:"S"`

	LastPrice string              `json:"c"`
	Ignore2   utils.NullJSONValue `json:"C"`

	OpenPrice string              `json:"o"`
	Ignore3   utils.NullJSONValue `json:"O"`

	HighPrice string              `json:"h"`
	Ignore4   utils.NullJSONValue `json:"H"`

	LowPrice string              `json:"l"`
	Ignore5  utils.NullJSONValue `json:"L"`

	BaseVolume string              `json:"v"`
	Ignore6    utils.NullJSONValue `json:"V"`

	QuoteVolume string              `json:"q"`
	Ignore7     utils.NullJSONValue `json:"Q"`
}

func mapWSAllMiniMarketsStatEvent(message []byte) (*WSAllMiniMarketsStatEvent, error) {
	var event *WSAllMiniMarketsStatEvent
	err := json.Unmarshal(message, &event)
	if err != nil {
		return nil, errors.Wrap(err, "can't unmarshal JSON")
	}
	return event, nil
}
