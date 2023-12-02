package krisa_phemex_fork

import (
	"context"
	"encoding/json"

	"github.com/Krisa/go-phemex/common"
)

// CancelOrderService cancel an order
type CancelOrderService struct {
	c       *Client
	symbol  string
	orderID *string
}

// Symbol set symbol
func (s *CancelOrderService) Symbol(symbol string) *CancelOrderService {
	s.symbol = symbol
	return s
}

// OrderID set orderID
func (s *CancelOrderService) OrderID(orderID string) *CancelOrderService {
	s.orderID = &orderID
	return s
}

// Do send request
// `rateLimHeaders` can be used <=> it isn't nil; despite the error
func (s *CancelOrderService) Do(ctx context.Context, opts ...RequestOption) (
	res *OrderResponse, rateLimHeaders *RateLimiterHeaders, err error,
) {
	r := &request{
		method:   "DELETE",
		endpoint: "/orders/cancel",
		secType:  secTypeSigned,
	}
	r.setParam("symbol", s.symbol)
	if s.orderID != nil {
		r.setParam("orderID", *s.orderID)
	}
	data, rateLimHeaders, err := s.c.callAPI(ctx, r, opts...)
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
