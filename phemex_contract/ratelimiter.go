package phemex_contract

import (
	"time"

	"github.com/aulaleslie/trade-exchanges/phemex_contract/krisa_phemex_fork"
	"github.com/aulaleslie/trade-exchanges/utils"
	"go.uber.org/zap"
)

type PhemexRateLimiter struct {
	// Limits are there https://github.com/phemex/phemex-api-docs/blob/master/Generic-API-Info.en.md
	// Also take a look to rate-limits.txt file
	Contract *PhemexGroupRateLimiter
	Other    *PhemexGroupRateLimiter
	Log      *zap.Logger
}

func NewPhemexRateLimiter(lg *zap.Logger) *PhemexRateLimiter {
	return &PhemexRateLimiter{
		Contract: NewPhemexGroupRateLimiter(500),
		Other:    NewPhemexGroupRateLimiter(100),
		Log:      lg.Named("PhemexRateLim"),
	}
}

// Can work in case of `rateLimHeaders == nil`
func (rl *PhemexRateLimiter) Apply(rateLimHeaders *krisa_phemex_fork.RateLimiterHeaders) {
	if rateLimHeaders == nil {
		return
	}

	switch rateLimHeaders.GroupName {
	case krisa_phemex_fork.ContractGroupName:
		rl.Contract.Apply(rateLimHeaders)
	case krisa_phemex_fork.OthersGroupName:
		rl.Other.Apply(rateLimHeaders)
	default:
		rl.Log.Warn("Came rate limiters headers with unknown group", zap.Any("group", rateLimHeaders.GroupName))
	}
}

/////

type PhemexGroupRateLimiter struct {
	Lim *utils.ChangeableMinuteRateLimiter
}

func NewPhemexGroupRateLimiter(capacityPerMinute uint) *PhemexGroupRateLimiter {
	// x-ratelimit-remaining-groupName   Remaining request permits in this minute
	// x-ratelimit-capacity-groupName    Request ratelimit capacity
	// x-ratelimit-retry-after-groupName Reset timeout in seconds for current ratelimited user
	return &PhemexGroupRateLimiter{
		Lim: utils.NewChangeableMinuteRateLimiter(int(capacityPerMinute)),
	}
}

func (pgrl *PhemexGroupRateLimiter) Apply(rateLimHeaders *krisa_phemex_fork.RateLimiterHeaders) {
	if rateLimHeaders.RetryAfter != nil {
		pgrl.Lim.SetNextDuration(*rateLimHeaders.RetryAfter + time.Second)
	}
	// IMPROVEMENT: set Remaining too. But I think that this will be an excess.
}
