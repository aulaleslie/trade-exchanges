package krisa_phemex_fork

import (
	"context"
	"encoding/json"

	"github.com/Krisa/go-phemex/common"
)

// CreateOrderService create order
type CreateOrderService struct {
	c                *Client
	symbol           string
	clOrdID          *string
	actionBy         *string
	side             SideType
	orderQty         *float64
	priceEp          *int64
	ordType          *OrderType
	stopPxEp         *int64
	timeInForce      *TimeInForceType
	reduceOnly       *bool
	closeOnTrigger   *bool
	takeProfitEp     *int64
	stopLossEp       *int64
	pegOffsetValueEp *int64
	triggerType      *TriggerType
	text             *string
	pegPriceType     *string
}

// Symbol set symbol
func (s *CreateOrderService) Symbol(symbol string) *CreateOrderService {
	s.symbol = symbol
	return s
}

// ClOrdID set clOrID
func (s *CreateOrderService) ClOrdID(clOrdID string) *CreateOrderService {
	s.clOrdID = &clOrdID
	return s
}

// ActionBy set actionBy
func (s *CreateOrderService) ActionBy(actionBy string) *CreateOrderService {
	s.actionBy = &actionBy
	return s
}

// Side set side
func (s *CreateOrderService) Side(side SideType) *CreateOrderService {
	s.side = side
	return s
}

// OrderQty set orderQty
func (s *CreateOrderService) OrderQty(orderQty float64) *CreateOrderService {
	s.orderQty = &orderQty
	return s
}

// PriceEp set priceEp
func (s *CreateOrderService) PriceEp(priceEp int64) *CreateOrderService {
	s.priceEp = &priceEp
	return s
}

// OrdType set ordType
func (s *CreateOrderService) OrdType(ordType OrderType) *CreateOrderService {
	s.ordType = &ordType
	return s
}

// StopPxEp set stopPxEp
func (s *CreateOrderService) StopPxEp(stopPxEp int64) *CreateOrderService {
	s.stopPxEp = &stopPxEp
	return s
}

// TimeInForce set timeInForce
func (s *CreateOrderService) TimeInForce(timeInForce TimeInForceType) *CreateOrderService {
	s.timeInForce = &timeInForce
	return s
}

// ReduceOnly set reduceOnly
func (s *CreateOrderService) ReduceOnly(reduceOnly bool) *CreateOrderService {
	s.reduceOnly = &reduceOnly
	return s
}

// CloseOnTrigger set closeOnTrigger
func (s *CreateOrderService) CloseOnTrigger(closeOnTrigger bool) *CreateOrderService {
	s.closeOnTrigger = &closeOnTrigger
	return s
}

// TakeProfitEp set takeProfitEp
func (s *CreateOrderService) TakeProfitEp(takeProfitEp int64) *CreateOrderService {
	s.takeProfitEp = &takeProfitEp
	return s
}

// StopLossEp set stopLossEp
func (s *CreateOrderService) StopLossEp(stopLossEp int64) *CreateOrderService {
	s.stopLossEp = &stopLossEp
	return s
}

// TriggerType set triggerType
func (s *CreateOrderService) TriggerType(triggerType TriggerType) *CreateOrderService {
	s.triggerType = &triggerType
	return s
}

// Text set text
func (s *CreateOrderService) Text(text string) *CreateOrderService {
	s.text = &text
	return s
}

// PegOffsetValueEp set pegOffsetValueEp
func (s *CreateOrderService) PegOffsetValueEp(pegOffsetValueEp int64) *CreateOrderService {
	s.pegOffsetValueEp = &pegOffsetValueEp
	return s
}

// PegPriceType set pegPriceType
func (s *CreateOrderService) PegPriceType(pegPriceType string) *CreateOrderService {
	s.pegPriceType = &pegPriceType
	return s
}

// `rateLimHeaders` can be used <=> it isn't nil; despite the error
func (s *CreateOrderService) createOrder(ctx context.Context, endpoint string, opts ...RequestOption) (
	data []byte, rateLimHeaders *RateLimiterHeaders, err error,
) {
	r := &request{
		method:   "POST",
		endpoint: endpoint,
		secType:  secTypeSigned,
	}
	m := params{
		"symbol": s.symbol,
		"side":   s.side,
	}
	if s.clOrdID != nil {
		m["clOrdID"] = *s.clOrdID
	}
	if s.orderQty != nil {
		m["orderQty"] = *s.orderQty
	}
	if s.actionBy != nil {
		m["actionBy"] = *s.actionBy
	}
	if s.priceEp != nil {
		m["priceEp"] = *s.priceEp
	}
	if s.ordType != nil {
		m["ordType"] = *s.ordType
	}
	if s.stopPxEp != nil {
		m["stopPxEp"] = *s.stopPxEp
	}
	if s.timeInForce != nil {
		m["timeInForce"] = *s.timeInForce
	}
	if s.reduceOnly != nil {
		m["reduceOnly"] = *s.reduceOnly
	}
	if s.closeOnTrigger != nil {
		m["closeOnTrigger"] = *s.closeOnTrigger
	}
	if s.takeProfitEp != nil {
		m["takeProfitEp"] = *s.takeProfitEp
	}
	if s.stopLossEp != nil {
		m["stopLossEp"] = *s.stopLossEp
	}
	if s.triggerType != nil {
		m["triggerType"] = *s.triggerType
	}
	if s.text != nil {
		m["text"] = *s.text
	}
	if s.pegPriceType != nil {
		m["pegPriceType"] = *s.pegPriceType
	}
	if s.pegOffsetValueEp != nil {
		m["pegOffsetValueEp"] = *s.pegOffsetValueEp
	}
	r.setFormParams(m)
	data, rateLimHeaders, err = s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return []byte{}, rateLimHeaders, err
	}
	return data, rateLimHeaders, nil
}

// Do send request
// `rateLimHeaders` can be used <=> it isn't nil; despite the error
func (s *CreateOrderService) Do(ctx context.Context, opts ...RequestOption) (
	res *OrderResponse, rateLimHeaders *RateLimiterHeaders, err error,
) {
	data, rateLimHeaders, err := s.createOrder(ctx, "/orders", opts...)
	if err != nil {
		return nil, rateLimHeaders, err
	}
	resp := new(BaseResponse)
	resp.Data = new(OrderResponse)
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, rateLimHeaders, err
	}
	if resp.Code > 0 {
		return nil, rateLimHeaders, &common.APIError{
			Code:    resp.Code,
			Message: resp.Msg,
		}
	}
	return resp.Data.(*OrderResponse), rateLimHeaders, nil
}

// OrderResponse define create order response
type OrderResponse struct {
	BizError       int             `json:"bizError"`
	OrderID        string          `json:"orderID"`
	ClOrdID        string          `json:"clOrdID"`
	Symbol         string          `json:"symbol"`
	Side           SideType        `json:"side"`
	ActionTimeNs   int64           `json:"actionTimeNs"`
	TransactTimeNs int64           `json:"transactTimeNs"`
	OrderType      OrderType       `json:"orderType"`
	PriceEp        int64           `json:"priceEp"`
	Price          float64         `json:"price"`
	OrderQty       float64         `json:"orderQty"`
	DisplayQty     float64         `json:"displayQty"`
	TimeInForce    TimeInForceType `json:"timeInForce"`
	ReduceOnly     bool            `json:"reduceOnly"`
	TakeProfitEp   int64           `json:"takeProfitEp"`
	TakeProfit     float64         `json:"takeProfit"`
	StopPxEp       int64           `json:"stopPxEp"`
	StopPx         float64         `json:"stopPx"`
	StopLossEp     int64           `json:"stopLossEp"`
	ClosedPnlEv    int64           `json:"closedPnlEv"`
	ClosedPnl      float64         `json:"closedPnl"`
	ClosedSize     float64         `json:"closedSize"`
	CumQty         float64         `json:"cumQty"`
	CumValueEv     int64           `json:"cumValueEv"`
	CumValue       float64         `json:"cumValue"`
	LeavesQty      float64         `json:"leavesQty"`
	LeavesValueEv  int64           `json:"leavesValueEv"`
	LeavesValue    float64         `json:"leavesValue"`
	StopLoss       float64         `json:"stopLoss"`
	StopDirection  string          `json:"stopDirection"`
	OrdStatus      string          `json:"ordStatus"`
	Trigger        string          `json:"trigger"`
}
