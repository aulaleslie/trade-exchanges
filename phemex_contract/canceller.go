package phemex_contract

import (
	"context"

	"github.com/Krisa/go-phemex/common"
	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/phemex_contract/krisa_phemex_fork"
	"github.com/aulaleslie/trade-exchanges/utils"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type OrderCanceller struct {
	client       *krisa_phemex_fork.Client
	orderFetcher *CombinedOrdersFetcher

	lim *PhemexRateLimiter
	lg  *zap.Logger
}

func NewOrderCanceller(
	client *krisa_phemex_fork.Client,
	cof *CombinedOrdersFetcher,
	lim *PhemexRateLimiter,
	lg *zap.Logger,
) *OrderCanceller {
	return &OrderCanceller{
		client:       client,
		orderFetcher: cof,

		lim: lim,
		lg:  lg.Named("OrderCanceller"),
	}
}

func (oc *OrderCanceller) CancelOrder(ctx context.Context, phemexSymbol, orderID string) error {
	err := oc.cancelOrderInternal(ctx, phemexSymbol, orderID)
	if errors.Is(err, exchanges.OrderNotFoundError) {
		return utils.ReplaceError(
			errors.New("can't check orders status (order can be removed)"), err)
	}
	return err
}

func (oc *OrderCanceller) cancelOrderInternal(ctx context.Context, phemexSymbol, orderID string) error {
	{
		// The problem is that Phemex removes order which is received two cancelation requests
		// So for 2nd and next orders it's better to make a precheck
		alreadyCanceled, err := oc.isOrderAlreadyCanceled(ctx, phemexSymbol, orderID)
		if err != nil {
			return errors.Wrap(err, "can't make order status pre-cancel check")
		}
		if alreadyCanceled {
			return nil
		}
	}

	err := oc.sendCancellationRequest(ctx, phemexSymbol, orderID)
	// log.Printf("sendCancellationRequest error: %v", err)
	if !errors.Is(err, exchanges.OrderNotFoundError) {
		return err
	}

	oc.lg.Debug("Checking order info by orderID")
	// OPTIMIZATION: we can fetch and cache list of last orders instead of
	// fetching every order directly.
	orderInfo, _, err := oc.orderFetcher.GetOrderInfoByOrderID(ctx, phemexSymbol, orderID)
	if err != nil {
		return errors.Wrap(err, "can't make order status post-cancel check")
	}

	return oc.mapCancellationStatusToError(orderInfo.Status)
}

func (oc *OrderCanceller) isOrderAlreadyCanceled(ctx context.Context, phemexSymbol, orderID string) (bool, error) {
	orderInfo, _, err := oc.orderFetcher.GetOrderInfoByOrderID(ctx, phemexSymbol, orderID)
	if err != nil {
		return false, err
	}

	switch orderInfo.Status {
	case exchanges.UnknownOST, exchanges.NewOST, exchanges.PartiallyFilledOST:
		return false, nil
	case exchanges.FilledOST:
		return false, exchanges.OrderExecutedError
	case exchanges.CanceledOST, exchanges.RejectedOST, exchanges.ExpiredOST:
		return true, nil
	default:
		// TODO: log error
		return false, nil // Just allowing to call cancel method
	}
}

func (oc *OrderCanceller) mapCancellationStatusToError(status exchanges.OrderStatusType) error {
	switch status {
	case exchanges.UnknownOST:
		return errors.New("order have Unknown status")
	case exchanges.NewOST, exchanges.PartiallyFilledOST:
		return errors.Errorf("order wasn't canceled (status=%v)", status)
	case exchanges.FilledOST:
		return exchanges.OrderExecutedError
	case exchanges.CanceledOST, exchanges.RejectedOST, exchanges.ExpiredOST:
		return nil
	default:
		return errors.Errorf("can't cancel order + order nave unknown status = %v on Phemex", status)
	}
}

func (oc *OrderCanceller) sendCancellationRequest(ctx context.Context, phemexSymbol, orderID string) error {
	// OPTIMIZTION: in case of order changing there is AmendOrder/ReplaceOrder method
	oc.lim.Contract.Lim.Wait()
	order, rateLimHeaders, err := oc.client.NewCancelOrderService().
		OrderID(orderID).
		Symbol(phemexSymbol).
		Do(ctx)
	oc.lim.Apply(rateLimHeaders)
	_ = order
	// log.Printf("[OC-tryToCancel] order: %v", order)
	if IsAPINotFoundError(err) {
		return utils.ReplaceError(exchanges.OrderNotFoundError, err)
	}
	return errors.Wrap(err, "can't make cancel request")
}

func IsAPINotFoundError(err error) bool {
	if err == nil {
		return false
	}

	apiErr, ok := err.(*common.APIError)
	if !ok {
		return false
	}

	return apiErr.Code == OrderNotFoundCode
}
