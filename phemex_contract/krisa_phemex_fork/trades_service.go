package krisa_phemex_fork

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Krisa/go-phemex/common"
	"github.com/aulaleslie/trade-exchanges/utils"
)

// symbol=<symbol>&start=<start>&end=<end>&limit=<limit>&offset=<offset>&withCount=<withCount>
type TradesService struct {
	c         *Client
	symbol    string     // String  - ? which symbol needs to query - Trading symbols
	start     *time.Time // Integer - ? start time range, Epoch millis
	end       *time.Time // Integer - ? end time range, Epoch millis
	offset    *int       // Integer - ? offset to resultset
	limit     *int       // Integer - ? limit of resultset
	withCount *int       // Integer - ? probably page size
}

// symbol - String - ? which symbol needs to query - Trading symbols
func (s *TradesService) Symbol(symbol string) *TradesService {
	s.symbol = symbol
	return s
}

// start - Integer - ? start time range, Epoch millis
func (s *TradesService) Start(start time.Time) *TradesService {
	s.start = &start
	return s
}

// end - Integer - ? end time range, Epoch millis
func (s *TradesService) End(end time.Time) *TradesService {
	s.end = &end
	return s
}

// offset - Integer - ? offset to resultset
func (s *TradesService) Offset(offset int) *TradesService {
	s.offset = &offset
	return s
}

// limit - Integer - ? limit of resultset
func (s *TradesService) Limit(limit int) *TradesService {
	s.limit = &limit
	return s
}

// Do send request
// `rateLimiterHeaders` can be used <=> it isn't nil; despite the error
func (s *TradesService) Do(ctx context.Context, opts ...RequestOption) (
	res map[string]interface{}, rateLimiterHeaders *RateLimiterHeaders, err error,
) {
	// https://github.com/phemex/phemex-api-docs/blob/master/Public-Contract-API-en.md#query-user-trade
	// GET /exchange/order/trade?
	//   symbol=<symbol>&
	//   start=<start>&
	//   end=<end>&
	//   limit=<limit>&
	//   offset=<offset>&
	//   withCount=<withCount>

	r := &request{
		method:   "GET",
		endpoint: "/exchange/order/trade",
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
	if s.withCount != nil {
		r.setParam("withCount", *s.withCount)
	}

	data, rateLimiterHeaders, err := s.c.callAPI(ctx, r, opts...)
	if err != nil {
		return nil, rateLimiterHeaders, err
	}

	resp := new(BaseResponse)
	resp.Data = map[string]interface{}{}

	err = json.Unmarshal(data, &resp)
	if err != nil {
		return nil, rateLimiterHeaders, err
	}
	if resp.Code > 0 {
		return nil, rateLimiterHeaders, &common.APIError{
			Code:    resp.Code,
			Message: resp.Msg,
		}
	}
	result := resp.Data.(map[string]interface{})
	return result, rateLimiterHeaders, nil
}
