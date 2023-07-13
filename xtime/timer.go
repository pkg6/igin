package xtime

import "time"

//_ = SetInterval(func(args ...any) {
//		fmt.Println(args)
//	}, 1000, "hello world")
//	_ = SetTimeout(func(args ...any) {
//		fmt.Println(args)
//	}, 1200, "I ran once")
//	time.Sleep(5 * time.Second)

// Timer represents a monitor/iterator that runs
// a function on a given interval.
type Timer struct {
	interval     time.Duration
	maxIntervals int
	iterations   int
	fn           func(args ...any)
	args         []any
	ticker       *time.Ticker
	process      chan struct{}
	running      bool
}

// SetInterval runs the specified function every `X` milliseconds, where
// `X` is the specified interval duration.
func SetInterval(fn func(args ...any), duration int, args ...any) *Timer {
	timer := &Timer{
		running:      false,
		interval:     time.Duration(duration) * time.Millisecond,
		fn:           fn,
		args:         args,
		maxIntervals: -1,
	}
	timer.Start()
	return timer
}

// SetTimeout runs the specified function after waiting the specified duration (defined in milliseconds)
func SetTimeout(fn func(args ...any), duration int, args ...any) *Timer {
	timer := &Timer{
		running:      false,
		interval:     time.Duration(duration) * time.Millisecond,
		fn:           fn,
		args:         args,
		maxIntervals: 0,
	}
	timer.Start()
	return timer
}

// Start the timer.
func (timer *Timer) Start() {
	if timer.running {
		return
	}
	if timer.process != nil {
		close(timer.process)
	}
	timer.running = true
	timer.process = make(chan struct{})
	timer.ticker = time.NewTicker(timer.interval)
	timer.iterations = 0
	go func(timer *Timer) {
		for {
			if timer.running {
				select {
				case <-timer.ticker.C:
					go func(timer *Timer) {
						if !timer.running {
							return
						}
						timer.iterations = timer.iterations + 1
						if timer.maxIntervals > 0 && timer.iterations > timer.maxIntervals {
							timer.Stop()
							return
						}
						timer.fn(timer.args...)
						if timer.maxIntervals == 0 {
							timer.Stop()
							return
						}
					}(timer)
				case <-timer.process:
					timer.Stop()
					return
				}
			} else {
				return
			}
		}
	}(timer)
}

// Stop the timer.
func (timer *Timer) Stop() {
	if timer.running {
		close(timer.process)
		timer.running = false
	}
}

// Count represents the number of times the interval has been processed.
func (timer *Timer) Count() int {
	return timer.iterations
}

// IsRunning returns a boolean flag indicating whether the timer is running or not.
func (timer *Timer) IsRunning() bool {
	if timer == nil {
		return false
	}
	return timer.running
}

// Reset the timer
func (timer *Timer) Reset() {
	timer.Stop()
	timer.Start()
}
