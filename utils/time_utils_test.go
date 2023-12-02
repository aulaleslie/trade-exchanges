package utils

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCancellableTimerWithoutStop(t *testing.T) {
	timer := NewCancellableTimer()
	begin := time.Now()
	timer.Start(time.Millisecond * 400)
	<-timer.C
	end := time.Now()
	nanos := end.UnixNano() - begin.UnixNano()
	millis := nanos / 1000 / 1000
	assert.Truef(t, math.Abs(float64(millis-400)) < 100.0, "Elapsed millis: %d", millis)
}

func TestCancellableTimerWithStop(t *testing.T) {
	timer := NewCancellableTimer()
	timer.Start(time.Millisecond * 100)
	// timer.Start(0)
	timer.Stop()
	assert.Never(t,
		func() bool { <-timer.C; return true },
		time.Second,
		time.Millisecond*100,
	)
}

func TestCancellableTimerRestart(t *testing.T) {
	timer := NewCancellableTimer()
	begin := time.Now()
	timer.Start(time.Second / 2)
	timer.Stop()
	timer.Start(time.Millisecond * 700)
	<-timer.C
	end := time.Now()
	nanos := end.UnixNano() - begin.UnixNano()
	millis := nanos / 1000 / 1000
	assert.Truef(t, math.Abs(float64(millis-700)) < 100.0, "Elapsed millis: %d", millis)
}

func TestCancellableTimerStopAfterStop(t *testing.T) {
	timer := NewCancellableTimer()
	assert.Eventually(t, func() bool {
		timer.Start(time.Second / 2)
		timer.Stop()
		timer.Stop()
		return true
	}, 2*time.Second, time.Millisecond*100)
}
