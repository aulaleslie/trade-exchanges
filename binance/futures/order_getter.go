package futures

import (
	"context"
	"fmt"
	"strconv"

	"github.com/adshao/go-binance/v2/common"
	api "github.com/adshao/go-binance/v2/futures"
	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/utils"
	"github.com/cockroachdb/apd"
	"github.com/pkg/errors"
)

type OrderGetter struct {
	client *api.Client
}

func NewOrderGetter(client *api.Client) *OrderGetter {
	return &OrderGetter{
		client: client,
	}
}

func (og *OrderGetter) GetOrderInfoByClientOrderID(ctx context.Context, symbol, clientOrderID string) (exchanges.OrderInfo, error) {
	orderInfo := exchanges.OrderInfo{}

	data, err := og.GetBinanceOrder(ctx, symbol, clientOrderID)
	if err != nil {
		return orderInfo, errors.Wrap(err, "can't query order")
	}

	orderInfo.ClientOrderID = &data.ClientOrderID
	orderInfo.ID = data.ClientOrderID
	orderInfo.Status = mapOrderStatusType(string(data.Status))
	return orderInfo, nil
}

func (og *OrderGetter) GetOrderBinanceStatus(ctx context.Context, symbol, clientOrderID string) (api.OrderStatusType, error) {
	order, err := og.GetBinanceOrder(ctx, symbol, clientOrderID)
	if err != nil {
		return "", err
	}

	return order.Status, nil
}

func (og *OrderGetter) GetBinanceOrder(ctx context.Context, symbol, clientOrderID string) (*api.Order, error) {
	order, err := og.client.NewGetOrderService().
		Symbol(symbol).
		OrigClientOrderID(clientOrderID).
		Do(ctx)
	if err != nil {
		if og.isNotFoundDuringGetOrderStatus(err) {
			return nil, utils.ReplaceError(exchanges.OrderNotFoundError, err)
		}
		return nil, err
	}

	return order, nil
}

func (og *OrderGetter) isNotFoundDuringGetOrderStatus(err error) bool {
	// https://binance-docs.github.io/apidocs/spot/en/#11xx-2xxx-request-issues
	if err == nil {
		return false
	}

	apiErr, ok := err.(*common.APIError)
	if !ok {
		return false
	}

	// NOTE: This code only corresponds to GetOrder method.
	return apiErr.Code == -2013
}

func (og *OrderGetter) getBinanceOpenOrder(ctx context.Context) ([]*api.Order, error) {
	openOrders, err := og.client.NewListOpenOrdersService().
		Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return openOrders, nil
}

func (og *OrderGetter) buildOrderDetailInfo(order *api.Order) (res exchanges.OrderDetailInfo, err error) {
	price, _, err := apd.NewFromString(order.Price)
	if err != nil {
		return res, err
	}

	quantity, _, err := apd.NewFromString(order.OrigQuantity)
	if err != nil {
		return res, err
	}

	executedQuantity, _, err := apd.NewFromString(order.ExecutedQuantity)
	if err != nil {
		return res, err
	}

	orderStatusType := mapOrderStatusType(string(order.Status))

	orderType := mapOrderType(string(order.Type))

	orderSide := mapOrderSide(string(order.Side))

	orderTimeInForce := mapOrderTimeInForce(string(order.TimeInForce))

	stopPrice, _, err := apd.NewFromString(order.StopPrice)
	if err != nil {
		return res, err
	}

	res = exchanges.OrderDetailInfo{
		Symbol:        order.Symbol,
		ID:            fmt.Sprintf("%v", order.OrderID),
		ClientOrderID: &order.ClientOrderID,
		Price:         price,
		Quantity:      quantity,
		ExecutedQty:   executedQuantity,
		Status:        orderStatusType,
		OrderType:     orderType,
		Time:          order.Time,
		OrderSide:     orderSide,
		TimeInForce:   orderTimeInForce,
		StopPrice:     stopPrice,
		QuoteQuantity: nil,
	}
	return
}

func (og *OrderGetter) GetOpenOrders(ctx context.Context) (res []exchanges.OrderDetailInfo, err error) {
	openOrders, err := og.getBinanceOpenOrder(ctx)
	if err != nil {
		err = errors.Wrap(err, "can't query open order")
		fmt.Println(err)
		return
	}
	for _, openOrder := range openOrders {
		orderDetailInfo, err := og.buildOrderDetailInfo(openOrder)
		if err != nil {
			return res, err
		}
		res = append(res, orderDetailInfo)
	}
	return
}

func (og *OrderGetter) GetHistoryOrders(
	ctx context.Context,
	symbol *string,
	orderID *string,
	clientOrderID *string,
) (res []exchanges.OrderDetailInfo, err error) {
	if orderID != nil || clientOrderID != nil {
		orderService := og.client.NewGetOrderService()
		if orderID != nil {
			orderIDInt, err := strconv.Atoi(*orderID)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			orderService.OrderID(int64(orderIDInt))
		}

		if clientOrderID != nil {
			orderService.OrigClientOrderID(*clientOrderID)
		}

		if symbol != nil {
			orderService.Symbol(*symbol)
		}

		order, err := orderService.Do(context.Background())
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		orderDetailInfo, err := og.buildOrderDetailInfo(order)
		if err != nil {
			return nil, err
		}

		res = append(res, orderDetailInfo)
		return res, err
	}

	orders, err := og.client.NewListOrdersService().Do(ctx)
	if err != nil {
		err = errors.Wrap(err, "can't query open order")
		fmt.Println(err)
		return nil, err
	}
	for _, order := range orders {
		orderDetailInfo, err := og.buildOrderDetailInfo(order)
		if err != nil {
			return res, err
		}
		res = append(res, orderDetailInfo)
	}
	return
}
