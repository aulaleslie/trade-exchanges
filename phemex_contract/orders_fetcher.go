package phemex_contract

import (
	"context"
	"time"

	"github.com/Krisa/go-phemex"
	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/phemex_contract/krisa_phemex_fork"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type OldOrdersFetcher struct {
	client     *phemex.Client
	forkClient *krisa_phemex_fork.Client
	// clientOrderIDToOrderIDCache map[ClientOrderIDSelector]string // TODO: GC or autoexpiration?
}

func NewOldOrdersFetcher(apiKey, secretKey string, lg *zap.Logger) *OldOrdersFetcher {
	return &OldOrdersFetcher{
		client:     phemex.NewClient(apiKey, secretKey),
		forkClient: krisa_phemex_fork.NewClient(apiKey, secretKey, lg.Named("OldOrderFetcher")),
		// clientOrderIDToOrderIDCache: map[ClientOrderIDSelector]string{},
	}
}

func (of *OldOrdersFetcher) xaddToCache(order *phemex.OrderResponse) {
	if order.ClOrdID == "" {
		return
	}

	// selector := ClientOrderIDSelector{
	// 	Symbol:        order.Symbol,
	// 	ClientOrderID: order.ClOrdID,
	// }
	// of.clientOrderIDToOrderIDCache[selector] = order.OrderID
}

func (of *OldOrdersFetcher) unifiedGetOrderInfo(
	ctx context.Context,
	svc *krisa_phemex_fork.QueryOrderService,
) (exchanges.OrderInfo, *krisa_phemex_fork.OrderResponse, error) {
	resp, _, err := svc.Do(ctx) // TODO: add rate limiter before use and apply the headers too.
	if IsAPINotFoundError(err) {
		return exchanges.OrderInfo{}, nil, errors.Wrap(err, "unexpected not found error")
	}
	if err != nil {
		return exchanges.OrderInfo{}, nil, errors.Wrap(err, "unable to send request")
	}

	if len(resp) == 0 {
		return exchanges.OrderInfo{}, nil, exchanges.OrderNotFoundError
	}
	if len(resp) > 1 {
		return exchanges.OrderInfo{}, nil, errors.Errorf("too many orders in response")
	}

	oInfo, err := convertOrder(resp[0])
	if err != nil {
		return exchanges.OrderInfo{}, nil, errors.Wrap(err, "can't convert order")
	}
	return oInfo, resp[0], nil
}

func (of *OldOrdersFetcher) GetOrderInfoByOrderID(
	ctx context.Context, symbol, id string,
) (exchanges.OrderInfo, *krisa_phemex_fork.OrderResponse, error) {
	return of.unifiedGetOrderInfo(
		ctx,
		of.forkClient.NewQueryOrderService().
			Symbol(symbol).
			OrderID(id))
}

func (of *OldOrdersFetcher) GetOrderInfoByClientOrderID(
	ctx context.Context, symbol, clientOrderID string, createdAt *time.Time,
) (exchanges.OrderInfo, *krisa_phemex_fork.OrderResponse, error) {
	if createdAt == nil {
		return exchanges.OrderInfo{}, nil, errors.New("`createdAt` field is required")
	}

	// Experimantally meaused threshold is about 7 days
	if time.Since(*createdAt) > 3*24*time.Hour {
		return exchanges.OrderInfo{}, nil, errors.New("the order was created too much ago")
	}

	return of.unifiedGetOrderInfo(
		ctx,
		of.forkClient.NewQueryOrderService().
			Symbol(symbol).
			ClOrderID(clientOrderID))
}

func (of *OldOrdersFetcher) xFullScanInactiveOrders(
	symbol, clientOrderID string, FromTime time.Time,
) (*exchanges.OrderInfo, error) {
	limit := 200 // Experimantally measured threshold
	var prevOrderIDs map[string]struct{}
	offset := 0
	for {
		list, _, err := of.forkClient.NewListInactiveOrdersService().
			Symbol(symbol).
			Start(FromTime).
			Limit(limit).
			Offset(offset).
			Do(context.TODO())
		if err != nil {
			return nil, err
		}

		containsPrevOrderID := prevOrderIDs == nil
		currentOrderIDs := map[string]struct{}{}
		for _, order := range list {
			// of.addToCache(order)
			currentOrderIDs[order.OrderID] = struct{}{}

			if _, ok := prevOrderIDs[order.OrderID]; ok {
				containsPrevOrderID = true
			}
		}

		for _, order := range list {
			if order.ClOrdID == clientOrderID && order.Symbol == symbol {
				result, err := convertOrder(order)
				return &result, err
			}
		}

		if !containsPrevOrderID {
			offset = offset - limit/2
			continue
		}

		if len(list) < limit {
			return nil, nil
		}

		offset = offset + limit - 5
		prevOrderIDs = currentOrderIDs
	}

}

func (of *OldOrdersFetcher) GetActiveOrders(symbol string) ([]*krisa_phemex_fork.OrderResponse, error) {
	orders, _, err := of.forkClient.NewListOpenOrdersService().Symbol(symbol).Do(context.TODO())
	return orders, err
}
