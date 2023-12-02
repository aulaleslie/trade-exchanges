package phemex_contract

import (
	"context"

	"github.com/Krisa/go-phemex/common"
	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/phemex_contract/krisa_phemex_fork"
	"github.com/aulaleslie/trade-exchanges/utils"
	"github.com/cockroachdb/apd"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type orderFields struct {
	ClOrdID     string
	OrdType     krisa_phemex_fork.OrderType
	OrderQty    float64
	PriceEp     int64
	Side        krisa_phemex_fork.SideType
	Symbol      string
	TimeInForce krisa_phemex_fork.TimeInForceType
}

func (of *orderFields) ToAPI(c *krisa_phemex_fork.Client) *krisa_phemex_fork.CreateOrderService {
	return c.NewCreateOrderService().
		// ActionBy()
		ClOrdID(of.ClOrdID).
		// CloseOnTrigger()
		OrdType(of.OrdType).
		OrderQty(of.OrderQty).
		// PegOffsetValueEp()
		// PegPriceType()
		PriceEp(of.PriceEp).
		// ReduceOnly() - ignore
		Side(of.Side).
		// StopLossEp()
		// StopPxEp()
		Symbol(of.Symbol).
		// TakeProfitEp()
		TimeInForce(of.TimeInForce)
	// TriggerType()
}

func (of *orderFields) Equal(x *orderResponse) bool {
	return of.ClOrdID == x.fields.ClOrdID &&
		of.OrdType == x.fields.OrdType &&
		of.OrderQty == x.fields.OrderQty &&
		of.PriceEp == x.fields.PriceEp &&
		of.Side == x.fields.Side &&
		of.Symbol == x.fields.Symbol &&
		of.TimeInForce == x.fields.TimeInForce
}

type OrderPlacer struct {
	orderGetter *CombinedOrdersFetcher
	client      *krisa_phemex_fork.Client

	lim *PhemexRateLimiter
	lg  *zap.Logger
}

func NewOrderPlacer(
	client *krisa_phemex_fork.Client,
	cof *CombinedOrdersFetcher,
	lim *PhemexRateLimiter,
	lg *zap.Logger,
) *OrderPlacer {
	return &OrderPlacer{
		client:      client,
		orderGetter: cof,

		lim: lim,
		lg:  lg.Named("OrderPlacer"),
	}
}

// Don't forget to floor `price` and `quantity`
func (op *OrderPlacer) CreateOrderRequest(
	symbol string, price, quantity *apd.Decimal, clientOrderID string, side krisa_phemex_fork.SideType,
) (*orderFields, error) {

	// In comparison with Binance
	// +Symbol:           symbol,
	// +Side:             side,
	// +Type:             api.OrderTypeLimit,
	// +TimeInForce:      api.TimeInForceTypeGTC,
	// +Quantity:         qtyFloored,
	// +Price:            priceFloored,
	// +NewClientOrderID: prefferedID,
	// NewOrderRespType: api.NewOrderRespTypeACK,

	qtyFloat64, err := utils.ToIntegerInFloat64(quantity)
	if err != nil {
		return nil, errors.Wrap(err, "can't convert quantity")
	}
	priceEp, _, err := ConvertPhemexPriceToPriceEp(symbol, price)
	if err != nil {
		return nil, errors.Wrap(err, "can't convert price")
	}

	// log.Printf("CreateOrderRequest: qtyFloat64=%v, priceEp=%v", qtyFloat64, priceEp)

	// TODO: write note about Stop field. Because we can use it to cancel bot orders more faster

	return &orderFields{
		ClOrdID:     clientOrderID,
		OrdType:     krisa_phemex_fork.OrderTypeLimit,
		OrderQty:    qtyFloat64,
		PriceEp:     priceEp,
		Side:        side,
		Symbol:      symbol,
		TimeInForce: krisa_phemex_fork.TimeInForceTypeGTC,
	}, nil
}

func (op *OrderPlacer) PlaceOrder(
	ctx context.Context,
	isRetry bool,
	symbol string, price, quantity *apd.Decimal, clientOrderID string,
	side krisa_phemex_fork.SideType,
) (id string, e error) {
	// log.Printf("Place Order: price=%v, quantity=%v", price, quantity)

	orderReq, ordReqErr := op.CreateOrderRequest(symbol, price, quantity, clientOrderID, side)
	if ordReqErr != nil {
		return "", errors.Wrapf(ordReqErr, "can't create order req")
	}

	// Client order ID isn't unique for Phemex! (found by experiment)

	if isRetry {
		// log.Printf("[OP] pre-check")
		id, preCheckErr := op.fetchAndCompare(ctx, orderReq)
		switch {
		case errors.Is(preCheckErr, exchanges.OrderNotFoundError):
			// ok
		case preCheckErr != nil:
			return "", errors.Wrap(preCheckErr, "first fetch and compare order")
		default:
			return id, nil
		}
	}

	// log.Printf("[OP] try to place")
	id, placeErr := op.tryToPlaceOrder(ctx, orderReq)
	if placeErr == nil {
		return id, nil
	}
	if !errors.Is(placeErr, exchanges.NewOrderRejectedError) {
		return "", placeErr
	}

	// log.Printf("[OP] post-check")
	id, getErr := op.fetchAndCompare(ctx, orderReq)
	switch {
	case errors.Is(getErr, exchanges.OrderNotFoundError) || errors.Is(getErr, exchanges.NewOrderRejectedError):
		return "", placeErr // Order rejected by another reason
	case getErr != nil:
		return "", errors.Wrapf(placeErr, "[Subreason: second fetch and compare order: %v]", getErr)
	default:
		return id, nil
	}
}

func (op *OrderPlacer) fetchAndCompare(ctx context.Context, req *orderFields) (id string, e error) {
	respOrderInfo, respOrder, getErr := op.orderGetter.GetOrderInfoByClientOrderID(ctx, req.Symbol, req.ClOrdID)
	if errors.Is(getErr, exchanges.OrderNotFoundError) {
		// log.Printf("[OP-fac] NotFound")
		return "", exchanges.OrderNotFoundError
	}
	if getErr != nil {
		// log.Printf("[OP-fac] getErr != nil")
		return "", errors.Wrapf(getErr, "can't fetch order")
	}
	// log.Printf("[OP] order fetched")

	if !req.Equal(respOrder) {
		op.lg.Warn("Order with same ClientOrderID is different", zap.Any("req", req), zap.Any("resp", respOrder))
		return "", errors.Errorf("different order with same ClientOrderID (%s) was placed", req.ClOrdID)
	}
	// log.Printf("[OP] same order")
	if respOrderInfo.Status == exchanges.RejectedOST {
		return "", exchanges.NewOrderRejectedError
	}
	// log.Printf("[OP] not rejected")
	return respOrderInfo.ID, nil
}

func (op *OrderPlacer) tryToPlaceOrder(ctx context.Context, req *orderFields) (id string, e error) {
	// log.Printf("tryToPlaceOrder: %v", req)
	op.lim.Contract.Lim.Wait()
	order, rateLimHeaders, err := req.ToAPI(op.client).Do(ctx)
	op.lim.Apply(rateLimHeaders)
	if op.isOrderRejectedError(err) {
		return "", utils.ReplaceError(exchanges.NewOrderRejectedError, err)
	}
	if err != nil {
		return "", errors.Wrapf(err, "can't place %s order", string(req.Side))
	}

	orderInfo, err := convertOrder(order)
	if err != nil {
		return "", errors.Wrap(err, "unable to convert status")
	}
	if orderInfo.Status == exchanges.RejectedOST {
		return "", errors.Wrap(
			exchanges.NewOrderRejectedError, "order have status = REJECTED")
	}

	return order.OrderID, nil
}

func (op *OrderPlacer) isOrderRejectedError(err error) bool {
	if err == nil {
		return false
	}

	apiErr, ok := err.(*common.APIError)
	if !ok {
		return false
	}

	return isOrderRejectedCode(apiErr.Code)
}

func (op *OrderPlacer) PlaceOrderV2(ctx context.Context, isRetry bool, symbol string, price,
	quantity *apd.Decimal, clientOrderID string,
	side krisa_phemex_fork.SideType, orderType krisa_phemex_fork.OrderType) (id string, e error) {

	orderReq, ordReqErr := op.CreateOrderRequestV2(symbol, price, quantity, clientOrderID, side, orderType)
	if ordReqErr != nil {
		return "", errors.Wrapf(ordReqErr, "can't create order req")
	}

	if isRetry {
		id, preCheckErr := op.fetchAndCompare(ctx, orderReq)
		switch {
		case errors.Is(preCheckErr, exchanges.OrderNotFoundError):
			// ok
		case preCheckErr != nil:
			return "", errors.Wrap(preCheckErr, "first fetch and compare order")
		default:
			return id, nil
		}
	}

	id, placeErr := op.tryToPlaceOrder(ctx, orderReq)
	if placeErr == nil {
		return id, nil
	}
	if !errors.Is(placeErr, exchanges.NewOrderRejectedError) {
		return "", placeErr
	}

	id, getErr := op.fetchAndCompare(ctx, orderReq)
	switch {
	case errors.Is(getErr, exchanges.OrderNotFoundError) || errors.Is(getErr, exchanges.NewOrderRejectedError):
		return "", placeErr
	case getErr != nil:
		return "", errors.Wrapf(placeErr, "[Subreason: second fetch and compare order: %v]", getErr)
	default:
		return id, nil
	}
}

func (op *OrderPlacer) CreateOrderRequestV2(symbol string, price,
	quantity *apd.Decimal, clientOrderID string,
	side krisa_phemex_fork.SideType,
	orderType krisa_phemex_fork.OrderType) (*orderFields, error) {

	qtyFloat64, err := utils.ToIntegerInFloat64(quantity)
	if err != nil {
		return nil, errors.Wrap(err, "can't convert quantity")
	}
	priceEp, _, err := ConvertPhemexPriceToPriceEp(symbol, price)
	if err != nil {
		return nil, errors.Wrap(err, "can't convert price")
	}

	return &orderFields{
		ClOrdID:     clientOrderID,
		OrdType:     orderType,
		OrderQty:    qtyFloat64,
		PriceEp:     priceEp,
		Side:        side,
		Symbol:      symbol,
		TimeInForce: krisa_phemex_fork.TimeInForceTypeGTC,
	}, nil
}
