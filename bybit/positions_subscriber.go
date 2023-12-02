package bybit

import (
	"context"
	"fmt"

	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/utils"
	"github.com/hirokisan/bybit/v2"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func SubscribeToPositions(ctx context.Context, wsClient *bybit.WebSocketClient, lg *zap.Logger, category bybit.CategoryV5) (<-chan exchanges.PositionEvent, error) {
	svc, err := wsClient.V5().Private()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create V5 service")
	}
	out := make(chan exchanges.PositionEvent, 100)

	err = svc.Subscribe()
	if err != nil {
		return nil, errors.Wrap(err, "unable to subscribe V5 service")
	}

	_, err = svc.SubscribePosition(func(response bybit.V5WebsocketPrivatePositionResponse) error {
		payloads := make([]*exchanges.PositionPayload, 0)
		for _, positionData := range response.Data {

			if string(positionData.Category) == string(category) {
				symbol := fmt.Sprintf("%v", positionData.Symbol)

				lg.Sugar().Infof("Get bybit position from websocket with symbol: %v",
					string(symbol),
				)
				positionPayload := &exchanges.PositionPayload{
					Symbol: symbol,
					Value:  utils.FromString(positionData.PositionBalance),
				}

				payloads = append(payloads, positionPayload)
				out <- exchanges.PositionEvent{
					Payload: payloads}
			}
		}
		return err
	})

	if err != nil {
		out <- exchanges.PositionEvent{DisconnectedWithErr: fmt.Errorf("connection error %v", err)}
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
