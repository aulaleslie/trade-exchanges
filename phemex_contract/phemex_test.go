package phemex_contract

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGetTradableSymbols(t *testing.T) {
	lg := zap.NewExample()
	ph := NewPhemexContract("", "", NewPhemexRateLimiter(lg), lg)
	symbols, err := ph.GetTradableSymbols(context.Background())
	assert.NoError(t, err)
	assert.NotEmpty(t, symbols)

	t.Logf("symbols: %v", symbols)
}
