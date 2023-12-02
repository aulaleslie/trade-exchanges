package utils

import (
	"sync"
	"time"

	"go.uber.org/atomic"
)

type CancellableTimer struct {
	isRunning *atomic.Bool
	stopMutex sync.Mutex
	timer     *time.Timer
	stop      chan struct{}
	stopped   chan struct{}
	C         chan struct{}
}

func NewCancellableTimer() *CancellableTimer {
	return &CancellableTimer{
		isRunning: atomic.NewBool(false),
		stop:      make(chan struct{}),
		stopped:   make(chan struct{}),
		C:         make(chan struct{}),
	}
}

func (t *CancellableTimer) Start(d time.Duration) {
	t.isRunning.Store(true)

	t.timer = time.NewTimer(d)
	go func() {
		select {
		case <-t.stop:
			if !t.timer.Stop() {
				<-t.timer.C
			}
			t.isRunning.Store(false)

			t.stopped <- struct{}{}
			return
		case <-t.timer.C:
			t.isRunning.Store(false)

			t.C <- struct{}{}
			return
		}
	}()
	return
}

// Stop can be called many times. Works on already stopeed timer.
func (t *CancellableTimer) Stop() {
	t.stopMutex.Lock()
	defer t.stopMutex.Unlock()
	if !t.isRunning.Load() {
		return
	}

	t.stop <- struct{}{}
	<-t.stopped
}

///////////////////////////

func TimeUNIXMillis(t time.Time) int64 {
	return t.UnixNano() / (1000 * 1000)
}

func TimePtr(t time.Time) *time.Time {
	return &t
}
