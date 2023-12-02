package tradovate

import (
	"context"
	"fmt"

	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/utils"
	"github.com/hirokisan/bybit/v2"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func SubscribeToPrices(ctx context.Context, wsClient *bybit.WebSocketClient, lg *zap.Logger, symbol string) (<-chan exchanges.PriceEvent, error) {
	svc, err := wsClient.V5().Public(bybit.CategoryV5Spot)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create V5 service")
	}
	out := make(chan exchanges.PriceEvent, 100)

	_, err = svc.SubscribeTicker(
		bybit.V5WebsocketPublicTickerParamKey{
			Symbol: bybit.SymbolV5(symbol),
		},
		func(response bybit.V5WebsocketPublicTickerResponse) error {
			priceData := response.Data.Spot

			if priceData.LastPrice != "" {
				price := utils.FromString(priceData.LastPrice)
				out <- exchanges.PriceEvent{
					Payload: price}
			}

			return err
		})

	if err != nil {
		out <- exchanges.PriceEvent{DisconnectedWithErr: fmt.Errorf("connection error %v", err)}
		defer close(out)

		// handle registration error
		lg.Sugar().Info("registration connection", err.Error())
	}

	errHandler := func(isWebsocketClosed bool, err error) {
		// Connection issue (timeout, etc.).
		// TODO: At this point, the connection is dead and you must handle the reconnection yourself
		lg.Sugar().Info("err handler called", err.Error())
	}

	go svc.Start(context.Background(), errHandler)

	return out, nil
}

func SubscribeToPricesInverse(ctx context.Context, wsClient *bybit.WebSocketClient, lg *zap.Logger, symbol string) (<-chan exchanges.PriceEvent, error) {
	svc, err := wsClient.V5().Public(bybit.CategoryV5Inverse)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create V5 service")
	}
	out := make(chan exchanges.PriceEvent, 100)

	_, err = svc.SubscribeTicker(
		bybit.V5WebsocketPublicTickerParamKey{
			Symbol: bybit.SymbolV5(symbol),
		},
		func(response bybit.V5WebsocketPublicTickerResponse) error {
			priceData := response.Data.LinearInverse

			if priceData.LastPrice != "" {
				price := utils.FromString(priceData.LastPrice)
				out <- exchanges.PriceEvent{
					Payload: price}
			}

			return err
		})

	if err != nil {
		out <- exchanges.PriceEvent{DisconnectedWithErr: fmt.Errorf("connection error %v", err)}
		defer close(out)

		// handle registration error
		lg.Sugar().Info("registration connection", err.Error())
	}

	errHandler := func(isWebsocketClosed bool, err error) {
		// Connection issue (timeout, etc.).
		// TODO: At this point, the connection is dead and you must handle the reconnection yourself
		lg.Sugar().Info("err handler called", err.Error())
	}

	go svc.Start(context.Background(), errHandler)

	return out, nil
}

func SubscribeToPricesLinear(ctx context.Context, wsClient *bybit.WebSocketClient, lg *zap.Logger, symbol string) (<-chan exchanges.PriceEvent, error) {
	svc, err := wsClient.V5().Public(bybit.CategoryV5Linear)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create V5 service")
	}
	out := make(chan exchanges.PriceEvent, 100)

	_, err = svc.SubscribeTicker(
		bybit.V5WebsocketPublicTickerParamKey{
			Symbol: bybit.SymbolV5(symbol),
		},
		func(response bybit.V5WebsocketPublicTickerResponse) error {
			priceData := response.Data.LinearInverse

			if priceData.LastPrice != "" {
				price := utils.FromString(priceData.LastPrice)
				out <- exchanges.PriceEvent{
					Payload: price}
			}

			return err
		})

	if err != nil {
		out <- exchanges.PriceEvent{DisconnectedWithErr: fmt.Errorf("connection error %v", err)}
		defer close(out)

		// handle registration error
		lg.Sugar().Info("registration connection", err.Error())
	}

	errHandler := func(isWebsocketClosed bool, err error) {
		// Connection issue (timeout, etc.).
		// TODO: At this point, the connection is dead and you must handle the reconnection yourself
		lg.Sugar().Info("err handler called", err.Error())
	}

	go svc.Start(context.Background(), errHandler)

	return out, nil
}
