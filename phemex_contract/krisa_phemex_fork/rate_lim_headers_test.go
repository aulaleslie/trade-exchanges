package krisa_phemex_fork

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimHeadersParse(t *testing.T) {
	{
		in := http.Header{}
		in.Set("X-RateLimit-Capacity-CoNTRaCT", "100")
		in.Set("X-RateLimit-Remaining-COnTRAcT", "70")
		in.Set("X-RateLimit-Retry-After-CONtRACt", "23")

		out, err := ParseRateLimiterHeaders(in)
		assert.NoError(t, err)
		assert.Equal(t,
			&RateLimiterHeaders{
				GroupName:  ContractGroupName,
				Capacity:   100,
				Remaining:  &[]int{70}[0],
				RetryAfter: &[]time.Duration{23 * time.Second}[0],
			},
			out,
		)
	}

	{
		in := http.Header{}
		in.Set("X-RateLimit-Capacity-OthErs", "1100")
		in.Set("X-RateLimit-Remaining-OthErs", "170")
		in.Set("X-RateLimit-Retry-After-OthErs", "123")

		out, err := ParseRateLimiterHeaders(in)
		assert.NoError(t, err)
		assert.Equal(t,
			&RateLimiterHeaders{
				GroupName:  OthersGroupName,
				Capacity:   1100,
				Remaining:  &[]int{170}[0],
				RetryAfter: &[]time.Duration{123 * time.Second}[0],
			},
			out,
		)
	}

	{
		in := http.Header{}
		in.Set("X-RateLimit-Capacity-Unknown", "100")
		in.Set("X-RateLimit-Remaining-Unknown", "70")
		in.Set("X-RateLimit-Retry-After-Unknown", "23")

		out, err := ParseRateLimiterHeaders(in)
		assert.NoError(t, err)
		assert.Nil(t, out)
	}
}
