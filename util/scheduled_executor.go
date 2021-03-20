package util

import (
	"sync"
	"time"
)

type ScheduledExecutor interface {
	Start()
	Stop()
}

type CustomTicker interface {
	Stop()
	Start()
	C() <-chan time.Time
}

type Option func(*scheduledExecutor)

type defaultTicker struct {
	*time.Ticker
	period time.Duration
}

func (d *defaultTicker) C() <-chan time.Time {
	return d.Ticker.C
}

func (d *defaultTicker) Start() {
	d.Ticker = time.NewTicker(d.period)
}

type scheduledExecutor struct {
	cb           func()
	isRunning    AtomicBool
	ticker       CustomTicker
	done         chan bool
	wg           sync.WaitGroup
	initialDelay time.Duration
}

func NewScheduledExecutor(period time.Duration, cb func()) ScheduledExecutor {
	return &scheduledExecutor{
		cb:           cb,
		isRunning:    NewAtomicBool(false),
		wg:           sync.WaitGroup{},
		initialDelay: period,
		ticker:       CustomTicker(&defaultTicker{period: period}),
	}
}

func NewScheduledExecutorWithOpts(period time.Duration, cb func(), options ...Option) ScheduledExecutor {
	s := &scheduledExecutor{
		cb:           cb,
		isRunning:    NewAtomicBool(false),
		wg:           sync.WaitGroup{},
		initialDelay: period,
		ticker:       CustomTicker(&defaultTicker{period: period}),
	}

	for _, opt := range options {
		opt(s)
	}

	return s
}

func WithInitialDelay(initialDelay time.Duration) Option {
	return func(s *scheduledExecutor) {
		s.initialDelay = initialDelay
	}
}

func WithCustomTicker(ticker CustomTicker) Option {
	return func(s *scheduledExecutor) {
		s.ticker = ticker
	}
}

func (s *scheduledExecutor) Start() {
	if s.isRunning.Set(true) {
		return
	}

	s.done = make(chan bool)
	s.wg.Add(1)

	go s.poll()
}

func (s *scheduledExecutor) Stop() {
	if !s.isRunning.Set(false) {
		return
	}

	s.done <- true
	s.wg.Wait()
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.ticker = nil
	s.done = nil
}

func (s *scheduledExecutor) poll() {
	defer s.wg.Done()

	// pause for initial delay
	if s.initialDelay > 0 {
		delay := time.NewTicker(s.initialDelay)
		select {
		case <-s.done:
			delay.Stop()
			return
		case <-delay.C:
			delay.Stop()
		}
	}

	// infinite loop for scheduled execution
	s.ticker.Start()
	for {
		s.cb()
		select {
		case <-s.done:
			return
		case <-s.ticker.C():
		}
	}
}
