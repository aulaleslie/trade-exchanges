package krisa_phemex_fork

import (
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// According to https://github.com/phemex/phemex-api-docs/blob/master/Generic-API-Info.en.md#api-ratelimit-rules

type RateLimiterGroupName string

const (
	ContractGroupName RateLimiterGroupName = "Contract" // 500/minutes
	OthersGroupName   RateLimiterGroupName = "Others"   // 100/minutes
	// SpotOrderGroupName = "SpotOrder" // trail, premium 500/minutes
)

type RateLimiterGroupHeaderNames struct {
	GroupName            RateLimiterGroupName
	RemainingHeaderName  string // Remaining request permits in this minute
	CapacityHeaderName   string // Request ratelimit capacity
	RetryAfterHeaderName string // Reset timeout in seconds for current ratelimited user
}

var ContractRateLimiterHeadersNames = RateLimiterGroupHeaderNames{
	GroupName:            ContractGroupName,
	RemainingHeaderName:  "X-RateLimit-Remaining-CONTRACT",
	CapacityHeaderName:   "X-RateLimit-Capacity-CONTRACT",
	RetryAfterHeaderName: "X-RateLimit-Retry-After-CONTRACT",
}

var OthersRateLimiterHeadersNames = RateLimiterGroupHeaderNames{
	GroupName:            OthersGroupName,
	RemainingHeaderName:  "X-RateLimit-Remaining-OTHERS",
	CapacityHeaderName:   "X-RateLimit-Capacity-OTHERS",
	RetryAfterHeaderName: "X-RateLimit-Retry-After-OTHERS",
}

type RateLimiterHeaders struct {
	GroupName RateLimiterGroupName

	Capacity   int            // Request ratelimit capacity
	Remaining  *int           // Remaining request permits in this minute
	RetryAfter *time.Duration // Reset timeout for current ratelimited user
}

func (rl *RateLimiterHeaders) parseByGroup(headers http.Header, names *RateLimiterGroupHeaderNames) error {
	rl.GroupName = names.GroupName

	capStr := headers.Get(names.CapacityHeaderName)
	capacity, err := strconv.Atoi(capStr)
	if err != nil {
		return errors.Wrap(err, "can't parse 'capacity'")
	}
	rl.Capacity = capacity

	remainingStr := headers.Get(names.RemainingHeaderName)
	if remainingStr != "" {
		remaining, err := strconv.Atoi(remainingStr)
		if err != nil {
			return errors.Wrap(err, "can't parse 'remaining'")
		}
		rl.Remaining = &remaining
	}

	retryAfterStr := headers.Get(names.RetryAfterHeaderName)
	if retryAfterStr != "" {
		retryAfterSecond, err := strconv.Atoi(retryAfterStr)
		if err != nil {
			return errors.Wrap(err, "can't parse 'retryAfter'")
		}
		retryAfter := time.Duration(retryAfterSecond) * time.Second
		rl.RetryAfter = &retryAfter
	}
	return nil
}

// Can return `nil` in case of no rate limiting headers
func ParseRateLimiterHeaders(headers http.Header) (*RateLimiterHeaders, error) {
	contractCap := headers.Get(ContractRateLimiterHeadersNames.CapacityHeaderName)
	if contractCap != "" {
		res := &RateLimiterHeaders{}
		err := res.parseByGroup(headers, &ContractRateLimiterHeadersNames)
		if err != nil {
			return nil, errors.Wrap(err, "can't parse contract group")
		}
		return res, nil
	}

	othersCap := headers.Get(OthersRateLimiterHeadersNames.CapacityHeaderName)
	if othersCap != "" {
		res := &RateLimiterHeaders{}
		err := res.parseByGroup(headers, &OthersRateLimiterHeadersNames)
		if err != nil {
			return nil, errors.Wrap(err, "can't parse others group")
		}
		return res, nil
	}

	return nil, nil
}
