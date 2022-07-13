/* For license and copyright information please see LEGAL file in repository */

package timer

import (
	"../protocol"
)

// Tick will send a signal on the t.Signal() channel after each tick on initialized Timer.
// The period of the ticks is specified by the duration arguments.
// The ticker will adjust the time interval or drop ticks to make up for slow receivers.
// The durations must be greater than zero; if not, Tick() will panic.
// Stop the ticker to release associated resources.
func (t *Timer) Tick(first, interval protocol.Duration, periodNumber int64) {
	if first < 1 || interval < 1 {
		panic("timer: non-positive interval to tick. period must be non-negative,")
	}
	t.period = int64(interval)
	t.periodNumber = periodNumber
	t.Start(first)
}

// Tick is a convenience wrapper for Timer.Tick() providing access to the ticking.
// Unlike After() that providing access to (<-chan struct{}),
// due to client need a way to shut it down the underlying
// Ticker to recovered by the garbage collector; to prevent **"leaks"**.
func Tick(first, interval protocol.Duration, periodNumber int64) *Timer {
	var timer Timer
	timer.Init(nil, nil)
	timer.Tick(first, interval, periodNumber)
	return &timer
}

// TickFunc or Schedule waits for the duration to elapse and then calls callback in each duration elapsed
// in its own goroutine. It returns a Timer that can be used to cancel the call using its Stop method.
// Schedule an execution at a given time, then once per interval. A typical use case is to execute code once per day at 12am.
func TickFunc(first, interval protocol.Duration, periodNumber int64, cb func(arg any), arg any) *Timer {
	var timer Timer
	timer.Init(callback(cb).concurrentRun, arg)
	timer.Tick(first, interval, periodNumber)
	return &timer
}