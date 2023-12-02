package krisa_phemex_fork

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Krisa/go-phemex/common"
	"github.com/aulaleslie/trade-exchanges/utils"
)

// ListOpenOrdersService list opened orders
type ListOpenOrdersService struct {
	c      *Client
	symbol string
}

// Symbol set symbol
func (s *ListOpenOrdersService) Symbol(symbol string) *ListOpenOrdersService {
	s.symbol = symbol
	return s
}

// Do send request
// `rateLimHeaders` can be used <=> it isn't nil; despite the error
func (s *ListOpenOrdersService) Do(ctx context.Context, opts ...RequestOption) (
	res []*OrderResponse, rateLimHeaders *RateLimiterHeaders, err error,
) {
	r := &request{
		method:   "GET",
		endpoint: "/orders/activeList",
		secType:  secTypeSigned,
	}
	if s.symbol != "" {
		r.setParam("symbol", s.symbol)
	}
	data, rateLimHeaders, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return []*OrderResponse{}, rateLimHeaders, err
	}

	resp := new(BaseResponse)
	resp.Data = new(RowsOrderResponse)

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
	rows := resp.Data.(*RowsOrderResponse)
	return rows.Rows, rateLimHeaders, nil
}

// RowsOrderResponse rows order response
type RowsOrderResponse struct {
	Rows []*OrderResponse `json:"rows"`
}

//////////////

// ListInactiveOrdersService list closed orders
type ListInactiveOrdersService struct {
	c      *Client
	symbol string     // String - which symbol needs to query - Trading symbols
	start  *time.Time // Integer - start time range, Epoch millis
	end    *time.Time // Integer - end time range, Epoch millis
	offset *int       // Integer - offset to resultset
	limit  *int       // Integer - limit of resultset
	// ordStatus string // String - order status list filter - New, PartiallyFilled, Untriggered, Filled, Canceled
}

// symbol - String - which symbol needs to query - Trading symbols
func (s *ListInactiveOrdersService) Symbol(symbol string) *ListInactiveOrdersService {
	s.symbol = symbol
	return s
}

// start - Integer - start time range, Epoch millis
func (s *ListInactiveOrdersService) Start(start time.Time) *ListInactiveOrdersService {
	s.start = &start
	return s
}

// end - Integer - end time range, Epoch millis
func (s *ListInactiveOrdersService) End(end time.Time) *ListInactiveOrdersService {
	s.end = &end
	return s
}

// offset - Integer - offset to resultset
func (s *ListInactiveOrdersService) Offset(offset int) *ListInactiveOrdersService {
	s.offset = &offset
	return s
}

// limit - Integer - limit of resultset
func (s *ListInactiveOrdersService) Limit(limit int) *ListInactiveOrdersService {
	s.limit = &limit
	return s
}

// Do send request
// `rateLimHeaders` can be used <=> it isn't nil; despite the error
func (s *ListInactiveOrdersService) Do(ctx context.Context, opts ...RequestOption) (
	res []*OrderResponse, rateLimHeaders *RateLimiterHeaders, err error,
) {
	// https://github.com/phemex/phemex-api-docs/blob/master/Public-Contract-API-en.md#query-closed-orders-by-symbol
	// GET /exchange/order/list?
	//   symbol=<symbol>&
	//   start=<start>&
	//   end=<end>&
	//   offset=<offset>&
	//   limit=<limit>&
	//   ordStatus=<ordStatus>&
	//   withCount=<withCount>

	r := &request{
		method:   "GET",
		endpoint: "/exchange/order/list",
		secType:  secTypeSigned,
	}
	if s.symbol != "" {
		r.setParam("symbol", s.symbol)
	}
	if s.start != nil {
		r.setParam("start", utils.TimeUNIXMillis(*s.start))
	}
	if s.end != nil {
		r.setParam("end", utils.TimeUNIXMillis(*s.end))
	}
	if s.offset != nil {
		r.setParam("offset", *s.offset)
	}
	if s.limit != nil {
		r.setParam("limit", *s.limit)
	}

	data, rateLimHeaders, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return []*OrderResponse{}, rateLimHeaders, err
	}

	resp := new(BaseResponse)
	resp.Data = new(RowsOrderResponse)

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
	rows := resp.Data.(*RowsOrderResponse)
	return rows.Rows, rateLimHeaders, nil
}
