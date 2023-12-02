package utils

import (
	"log"
	"sync"
	"time"

	"go.uber.org/atomic"
	// "github.com/sirupsen/logrus"
)

type MinuteRateLimiter struct {
	maxPerMinute        int
	lock                *sync.Mutex
	count               int
	lastPeriodTruncated time.Time
}

func NewMinuteRateLimiter(maxPerMinute int) *MinuteRateLimiter {
	return &MinuteRateLimiter{
		maxPerMinute: maxPerMinute,
		lock:         &sync.Mutex{},
	}
}

func (r *MinuteRateLimiter) Wait() {
	r.lock.Lock()
	defer r.lock.Unlock()

	currentTime := time.Now()
	currentTimeTruncated := currentTime.Truncate(time.Minute)

	setOneReq := func() {
		r.lastPeriodTruncated = currentTimeTruncated
		r.count = 1
	}

	// this will be the first request
	if r.lastPeriodTruncated.IsZero() {
		setOneReq()
		return
	}

	// the minute is up so the timer can be reset
	if currentTimeTruncated.After(r.lastPeriodTruncated) {
		setOneReq()
		return
	}

	r.count++

	// this request is the last that can be done this minute, sleep until the minute is up
	if r.count > r.maxPerMinute {
		nextPeriod := currentTimeTruncated.Add(time.Minute)
		waitDuration := nextPeriod.Sub(currentTime)
		// logrus.Debugf("ratelimiter: sleeping %s", waitDuration.String())
		time.Sleep(waitDuration)

		setOneReq()
	}
}

//////////

type RemainingForMinute struct {
	remaining int
	minute    time.Time
}

type ChangeableMinuteRateLimiter struct {
	maxPerMinute int

	lock                *sync.Mutex
	remaining           int
	lastPeriodTruncated time.Time

	nextRemaining      atomic.Value // RemainingForMinute
	nextWaitingEndTime atomic.Value // time.Time
}

func NewChangeableMinuteRateLimiter(maxPerMinute int) *ChangeableMinuteRateLimiter {
	return &ChangeableMinuteRateLimiter{
		maxPerMinute: maxPerMinute,
		lock:         &sync.Mutex{},
		remaining:    maxPerMinute,
	}
}

func (r *ChangeableMinuteRateLimiter) SetRemaining(at time.Time, remaining uint) {
	atTruncated := at.Truncate(time.Minute)
	if at.Sub(atTruncated) < 3*time.Second {
		// this can be data for previous minute
		return
	}

	nextRemaining := RemainingForMinute{
		remaining: int(remaining),
		minute:    atTruncated,
	}
	r.nextRemaining.Store(nextRemaining)
}

func (r *ChangeableMinuteRateLimiter) SetNextDuration(d time.Duration) {
	nextWaitingEndTime := time.Now().Add(d)
	r.nextWaitingEndTime.Store(nextWaitingEndTime)
}

func (r *ChangeableMinuteRateLimiter) Wait() {
	r.lock.Lock()
	defer r.lock.Unlock()

	currentTime := time.Now()
	currentTimeTruncated := currentTime.Truncate(time.Minute)

	getRemaining := func() (x int) {
		defer func() { log.Printf("giveRemaining = %d", x) }()
		nextRemainingRef := r.nextRemaining.Load()
		if nextRemainingRef == nil {
			return r.remaining
		}
		nextRemaining := nextRemainingRef.(RemainingForMinute)

		nowTruncated := time.Now().Truncate(time.Minute)
		if !nextRemaining.minute.Equal(nowTruncated) {
			return r.remaining
		}

		return MinInt(r.remaining, nextRemaining.remaining)
	}
	setOneReq := func() {
		r.lastPeriodTruncated = currentTimeTruncated
		r.remaining = r.maxPerMinute - 1
	}
	waitForNextWaitingEndTime := func() {
		now := time.Now()
		remainingWaitingTimeRef := r.nextWaitingEndTime.Load()
		if remainingWaitingTimeRef == nil {
			return
		}
		remainingWaitingTime := remainingWaitingTimeRef.(time.Time)

		if now.Before(remainingWaitingTime) {
			time.Sleep(remainingWaitingTime.Sub(now))
		}
	}

	// this will be the first request
	if r.lastPeriodTruncated.IsZero() {
		setOneReq()
		waitForNextWaitingEndTime()
		return
	}

	// the minute is up so the timer can be reset
	if currentTimeTruncated.After(r.lastPeriodTruncated) {
		setOneReq()
		waitForNextWaitingEndTime()
		return
	}

	r.remaining = getRemaining() - 1

	// this request is the last that can be done this minute, sleep until the minute is up
	if r.remaining < 0 {
		nextPeriod := currentTimeTruncated.Add(time.Minute)
		waitDuration := nextPeriod.Sub(currentTime)
		// logrus.Debugf("ratelimiter: sleeping %s", waitDuration.String())
		time.Sleep(waitDuration)

		setOneReq()
	}
	waitForNextWaitingEndTime()
}
