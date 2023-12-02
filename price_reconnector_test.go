package exchanges

import (
	"context"
	"testing"

	"github.com/aulaleslie/trade-exchanges/utils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/atomic"
	"go.uber.org/zap"
)

func TestPriceReconnector(t *testing.T) {
	iter := atomic.NewUint32(1)
	connect := func(c context.Context) (<-chan PriceEvent, error) {
		if iter.Load() > 2 {
			return nil, errors.New("immediate exit")
		}
		ch := make(chan PriceEvent, 100)
		go func() {
			defer close(ch)
			for i := 0; i < 3; i++ {
				ch <- PriceEvent{
					Payload: utils.FromUint(uint(iter.Load())),
				}
			}
			ch <- PriceEvent{DisconnectedWithErr: errors.New("test disconnect")}
			iter.Inc()
		}()
		return ch, nil
	}
	per := NewPriceEventReconnector(connect, nil, zap.NewExample())
	out, err := per.Watch(context.Background())
	assert.NoError(t, err)
	lastPrice := utils.Zero
	for ev := range out {
		if ev.Payload != nil {
			lastPrice = ev.Payload
		}
		t.Log(ev)
	}
	assert.Equal(t, utils.FromUint(2), lastPrice)
}
