package bybit

import (
	"context"
	"fmt"

	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/hirokisan/bybit/v2"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func SubscribeToOrders(ctx context.Context, wsClient *bybit.WebSocketClient, lg *zap.Logger, category bybit.CategoryV5) (<-chan exchanges.OrderEvent, error) {
	svc, err := wsClient.V5().Private()
	if err != nil {
		return nil, errors.Wrap(err, "unable to create V5 service")
	}
	out := make(chan exchanges.OrderEvent, 100)

	err = svc.Subscribe()
	if err != nil {
		return nil, errors.Wrap(err, "unable to subscribe V5 service")
	}

	_, err = svc.SubscribeOrder(func(response bybit.V5WebsocketPrivateOrderResponse) error {
		for _, orderData := range response.Data {

			if orderData.Category == string(category) {
				orderID := fmt.Sprintf("%v", orderData.OrderID)
				symbol := ToBybitFullSymbol(string(orderData.Symbol))

				if orderData.OrderLinkID != "" {
					orderID = orderData.OrderLinkID
				}

				orderStatus := mapOrderStatusType(
					string(orderData.OrderStatus))
				lg.Sugar().Infof("Get bybit order from websocket with orderID: %v clientOrderID: %v status: %v symbol: %v",
					orderID,
					string(orderData.OrderLinkID),
					string(orderData.OrderStatus),
					string(symbol),
				)
				out <- exchanges.OrderEvent{Payload: &exchanges.OrderEventPayload{
					OrderID:     orderID,
					OrderStatus: orderStatus,
					Symbol:      &symbol,
				}}
			}
		}
		return err
	})

	if err != nil {
		out <- exchanges.OrderEvent{DisconnectedWithErr: fmt.Errorf("connection error %v", err)}
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

	lg.Sugar().Info("leaving the Subscribe to Order")
	return out, nil
}
