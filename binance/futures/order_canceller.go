package futures

import (
	"context"

	"github.com/adshao/go-binance/v2/common"
	api "github.com/adshao/go-binance/v2/futures"
	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/utils"
	"github.com/pkg/errors"
)

type BinanceOrderCanceller struct {
	client      *api.Client
	orderGetter *OrderGetter
}

func NewBinanceOrderCanceller(client *api.Client) *BinanceOrderCanceller {
	return &BinanceOrderCanceller{
		client:      client,
		orderGetter: &OrderGetter{client: client},
	}
}

func (b *BinanceOrderCanceller) CancelOrder(ctx context.Context, symbol, clientOrderID string) error {
	err := b.tryToCancel(ctx, symbol, clientOrderID)
	switch {
	case err == nil:
		return nil
	case !errors.Is(err, exchanges.OrderNotFoundError):
		return err
	}

	// OPTIMIZATION: we can fetch and cache list of last orders instead of
	// fetching every order directly.
	status, err := b.orderGetter.GetOrderBinanceStatus(ctx, symbol, clientOrderID)
	if err != nil {
		return err
	}

	switch status {
	case api.OrderStatusTypeNew, api.OrderStatusTypePartiallyFilled:
		return errors.Errorf("Can't cancel order + order nave status = %v on Binance", status)
	case api.OrderStatusTypeFilled:
		return exchanges.OrderExecutedError
	case api.OrderStatusTypeCanceled:
		return nil
	case api.OrderStatusTypeRejected:
		return nil
	case api.OrderStatusTypeExpired:
		return nil
	default:
		return errors.Errorf("Can't cancel order + order nave unknown status = %v on Binance", status)
	}
}

func (b *BinanceOrderCanceller) tryToCancel(ctx context.Context, symbol, clientOrderID string) error {
	result, err := b.client.NewCancelOrderService().
		Symbol(symbol).
		OrigClientOrderID(clientOrderID).
		Do(ctx)
	if b.isNotFoundDuringCancellation(err) {
		return utils.ReplaceError(exchanges.OrderNotFoundError, err)
	}
	if err != nil {
		return err
	}

	if result.Status != api.OrderStatusTypeCanceled {
		return errors.Errorf("Bad cancellation result status: %v", result.Status)
	}

	return nil
}

func (b *BinanceOrderCanceller) isNotFoundDuringCancellation(err error) bool {
	if err == nil {
		return false
	}

	apiErr, ok := err.(*common.APIError)
	if !ok {
		return false
	}

	return apiErr.Code == -2011 && apiErr.Message == "Unknown order sent."
}
