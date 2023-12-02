package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMinuteRateLimiter(t *testing.T) {
	assert.Eventually(t,
		func() bool {
			// For with sleep is to mitigate matching the minutes end and minute start between two waits
			for i := 0; i < 3; i++ {
				rl := NewMinuteRateLimiter(1)
				rl.Wait()

				rl = NewMinuteRateLimiter(2)
				rl.Wait()
				rl.Wait()

				time.Sleep(time.Millisecond * 10)
			}
			return true
		},
		time.Second/5,
		time.Millisecond*10,
	)

	assert.Never(t,
		func() bool {
			// For with sleep is to mitigate matching the minutes end and minute start between two waits
			for i := 0; i < 3; i++ {
				rl := NewMinuteRateLimiter(1)
				rl.Wait()
				rl.Wait()
				time.Sleep(time.Millisecond * 10)
			}
			return true
		},
		time.Second/5,
		time.Millisecond*10,
	)
}

func TestChangeableMinuteRateLimiter(t *testing.T) {
	assert.Eventually(t,
		func() bool {
			// For with sleep is to mitigate matching the minutes end and minute start between two waits
			for i := 0; i < 3; i++ {
				rl := NewChangeableMinuteRateLimiter(1)
				rl.Wait()

				rl = NewChangeableMinuteRateLimiter(2)
				rl.Wait()
				rl.Wait()

				time.Sleep(time.Millisecond * 10)
			}
			return true
		},
		time.Second/5,
		time.Millisecond*10,
	)

	assert.Never(t,
		func() bool {
			now := time.Now()
			if now.Sub(now.Truncate(time.Minute)) > 59*1800*time.Millisecond {
				time.Sleep(time.Second)
			}

			rl := NewChangeableMinuteRateLimiter(5)
			rl.Wait()
			rl.SetRemaining(time.Now(), 0)
			rl.Wait()
			return true
		},
		time.Second/5,
		time.Millisecond*10,
	)

	assert.Never(t,
		func() bool {
			rl := NewChangeableMinuteRateLimiter(5)
			rl.Wait()
			rl.SetNextDuration(time.Second)
			rl.Wait()
			return true
		},
		time.Second/5,
		time.Millisecond*10,
	)

	assert.Eventually(t,
		func() bool {
			now := time.Now()
			if now.Sub(now.Truncate(time.Minute)) > 59*1800*time.Millisecond {
				time.Sleep(time.Second)
			}

			rl := NewChangeableMinuteRateLimiter(5)
			rl.Wait()
			rl.SetNextDuration(time.Second / 10)
			rl.Wait()
			return true
		},
		time.Second/5,
		time.Millisecond*10,
	)

	assert.Never(t,
		func() bool {
			// For with sleep is to mitigate matching the minutes end and minute start between two waits
			for i := 0; i < 3; i++ {
				rl := NewChangeableMinuteRateLimiter(1)
				rl.Wait()
				rl.Wait()
				time.Sleep(time.Millisecond * 10)
			}
			return true
		},
		time.Second/5,
		time.Millisecond*10,
	)
}
