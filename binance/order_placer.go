package binance

import (
	"context"
	"strings"

	api "github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/common"
	exchanges "github.com/aulaleslie/trade-exchanges"
	"github.com/aulaleslie/trade-exchanges/utils"
	"github.com/cockroachdb/apd"
	"github.com/pkg/errors"
)

type orderFields struct {
	Symbol           string
	Side             api.SideType
	Type             api.OrderType
	TimeInForce      api.TimeInForceType
	Quantity         string
	Price            string
	NewClientOrderID string
	NewOrderRespType api.NewOrderRespType
}

func (of *orderFields) ToAPI(c *api.Client) *api.CreateOrderService {
	if of.Type == api.OrderTypeMarket {
		return c.
			NewCreateOrderService().
			Symbol(of.Symbol).
			Side(of.Side).
			Type(of.Type).
			Quantity(of.Quantity).
			NewClientOrderID(of.NewClientOrderID).
			NewOrderRespType(of.NewOrderRespType)
	} else {
		return c.
			NewCreateOrderService().
			Symbol(of.Symbol).
			Side(of.Side).
			Type(of.Type).
			TimeInForce(of.TimeInForce).
			Quantity(of.Quantity).
			Price(of.Price).
			NewClientOrderID(of.NewClientOrderID).
			NewOrderRespType(of.NewOrderRespType)
	}
}

func (of *orderFields) equalStringNumber(x string, y string) bool {
	xx, _, err := apd.NewFromString(x)
	if err != nil {
		return false
	}

	yy, _, err := apd.NewFromString(y)
	if err != nil {
		return false
	}

	return xx.Cmp(yy) == 0
}

func (of *orderFields) Equal(x *api.Order) bool {
	eq := of.equalStringNumber
	return (true &&
		of.Symbol == x.Symbol &&
		of.Side == x.Side &&
		of.Type == x.Type &&
		of.TimeInForce == x.TimeInForce &&
		eq(of.Quantity, x.OrigQuantity) &&
		eq(of.Price, x.Price) &&
		of.NewClientOrderID == x.ClientOrderID)
	// of.NewOrderRespType can be ignored
}

type OrderPlacer struct {
	orderGetter *OrderGetter
	client      *api.Client
}

func NewOrderPlacer(client *api.Client) *OrderPlacer {
	return &OrderPlacer{
		client:      client,
		orderGetter: &OrderGetter{client: client},
	}
}

// CreateOrderRequest Don't forget to floor `price` and `quantity`
func (op *OrderPlacer) CreateOrderRequest(
	symbol string, price, quantity *apd.Decimal, prefferedID string,
	side api.SideType,
) (*orderFields, error) {
	req := &orderFields{
		Symbol:           symbol,
		Side:             side,
		Type:             api.OrderTypeLimit,
		TimeInForce:      api.TimeInForceTypeGTC,
		Quantity:         utils.ToFlatString(quantity),
		Price:            utils.ToFlatString(price),
		NewClientOrderID: prefferedID,
		NewOrderRespType: api.NewOrderRespTypeACK,
	}
	return req, nil
}

func (op *OrderPlacer) PlaceOrder(ctx context.Context,
	symbol string, price, quantity *apd.Decimal, prefferedID string,
	side api.SideType,
) (id string, e error) {
	orderReq, err := op.CreateOrderRequest(symbol, price, quantity, prefferedID, side)
	if err != nil {
		return "", errors.Wrapf(err, "can't create order req")
	}

	id, placeErr := op.tryToPlaceOrder(ctx, orderReq)
	if placeErr == nil {
		return id, nil
	}
	if !errors.Is(placeErr, exchanges.NewOrderRejectedError) {
		return "", placeErr
	}

	binanceOrder, getErr := op.orderGetter.GetBinanceOrder(ctx, symbol, prefferedID)
	if errors.Is(getErr, exchanges.OrderNotFoundError) {
		return "", placeErr // Order rejected by another reason
	}
	if getErr != nil {
		return "", errors.Wrapf(placeErr, "[Subreason: can't fetch order: %v]", getErr)
	}

	if orderReq.Equal(binanceOrder) {
		return binanceOrder.ClientOrderID, nil
	}
	// TODO: check status there?
	return "", errors.Errorf("different order with same ClientOrderID (%s) was placed", prefferedID)
}

func (op *OrderPlacer) PlaceOrderV2(ctx context.Context, symbol string, price, quantity *apd.Decimal, preferredID string,
	side api.SideType, orderType api.OrderType) (id string, e error) {
	orderReq, err := op.CreateOrderRequestV2(symbol, price, quantity, preferredID, side, orderType)
	if err != nil {
		return "", errors.Wrapf(err, "can't create order req")
	}

	id, placeErr := op.tryToPlaceOrder(ctx, orderReq)
	if placeErr == nil {
		return id, nil
	}
	if !errors.Is(placeErr, exchanges.NewOrderRejectedError) {
		return "", placeErr
	}

	binanceOrder, getErr := op.orderGetter.GetBinanceOrder(ctx, symbol, preferredID)
	if errors.Is(getErr, exchanges.OrderNotFoundError) {
		return "", placeErr // Order rejected by another reason
	}
	if getErr != nil {
		return "", errors.Wrapf(placeErr, "[Subreason: can't fetch order: %v]", getErr)
	}

	if orderReq.Equal(binanceOrder) {
		return binanceOrder.ClientOrderID, nil
	}
	// TODO: check status there?
	return "", errors.Errorf("different order with same ClientOrderID (%s) was placed", preferredID)
}

func (op *OrderPlacer) tryToPlaceOrder(ctx context.Context, req *orderFields) (id string, e error) {
	order, err := req.ToAPI(op.client).Do(ctx)
	if err != nil {
		err = op.castToOrderRejecterErrorIfCan(err)
		return "", errors.Wrapf(err, "can't place %s order", string(req.Side))
	}

	if order.Status == api.OrderStatusTypeRejected {
		return "", errors.Wrap(
			exchanges.NewOrderRejectedError, "order have status = REJECTED")
	}

	return order.ClientOrderID, nil
}

func (op *OrderPlacer) TestOrder(ctx context.Context,
	symbol string, price, quantity *apd.Decimal, prefferedID string,
	side api.SideType,
) error {
	req, err := op.CreateOrderRequest(symbol, price, quantity, prefferedID, side)
	if err != nil {
		return errors.Wrap(err, "can't create order request")
	}

	err = req.ToAPI(op.client).Test(ctx)
	if err != nil {
		err = op.castToOrderRejecterErrorIfCan(err)
		return errors.Wrapf(err, "can't test %s order", string(side))
	}

	return nil
}

func (op *OrderPlacer) castToOrderRejecterErrorIfCan(err error) error {
	if err == nil {
		return nil
	}

	if op.isOrderRejectedError(err) {
		return utils.ReplaceError(exchanges.NewOrderRejectedError, err)
	}

	return err
}

func (op *OrderPlacer) isOrderRejectedError(err error) bool {
	if err == nil {
		return false
	}

	apiErr, ok := err.(*common.APIError)
	if !ok {
		return false
	}

	// Based on https://binance-docs.github.io/apidocs/spot/en/#11xx-2xxx-request-issues
	if apiErr.Code == -2010 {
		return true
	}

	// Based on real example: `<APIError> code=-1013, msg=Filter failure: PERCENT_PRICE`
	// And on https://binance-docs.github.io/apidocs/spot/en/#9xxx-filter-failures
	if apiErr.Code == -1013 && strings.HasPrefix(apiErr.Message, "Filter failure:") {
		return true
	}

	return false
}

func (op *OrderPlacer) CreateOrderRequestV2(symbol string, price, quantity *apd.Decimal, preferredID string, side api.SideType, orderType api.OrderType) (*orderFields, error) {
	req := &orderFields{
		Symbol:           symbol,
		Side:             side,
		Type:             orderType,
		TimeInForce:      api.TimeInForceTypeGTC,
		Quantity:         utils.ToFlatString(quantity),
		Price:            utils.ToFlatString(price),
		NewClientOrderID: preferredID,
		NewOrderRespType: api.NewOrderRespTypeACK,
	}
	return req, nil
}
