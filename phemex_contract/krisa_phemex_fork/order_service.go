package krisa_phemex_fork

import (
	"context"
	"encoding/json"

	"github.com/Krisa/go-phemex/common"
)

// QueryOrderService cancel an order
type QueryOrderService struct {
	c         *Client
	symbol    string
	orderID   *string
	clOrderID *string
}

// Symbol set symbol
func (s *QueryOrderService) Symbol(symbol string) *QueryOrderService {
	s.symbol = symbol
	return s
}

// OrderID set orderID
func (s *QueryOrderService) OrderID(orderID string) *QueryOrderService {
	s.orderID = &orderID
	return s
}

// ClOrderID set clOrderID
func (s *QueryOrderService) ClOrderID(clOrderID string) *QueryOrderService {
	s.clOrderID = &clOrderID
	return s
}

// Do send request
// `rateLimiterHeaders` can be used <=> it isn't nil; despite the error
func (s *QueryOrderService) Do(ctx context.Context, opts ...RequestOption) (res []*OrderResponse, rateLimHeaders *RateLimiterHeaders, err error) {
	r := &request{
		method:   "GET",
		endpoint: "/exchange/order",
		secType:  secTypeSigned,
	}
	r.setParam("symbol", s.symbol)
	if s.orderID != nil {
		r.setParam("orderID", *s.orderID)
	}
	if s.clOrderID != nil {
		r.setParam("clOrdID", *s.clOrderID)
	}
	data, rateLimHeaders, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, rateLimHeaders, err
	}

	resp := new(BaseResponse)
	resp.Data = new([]*OrderResponse)
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
	// log.Printf("Data: %v", string(data)) // TODO: remove
	return *resp.Data.(*[]*OrderResponse), rateLimHeaders, nil
}
