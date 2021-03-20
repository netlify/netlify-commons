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
	C() <-chan time.Time
}

type Option func(*scheduledExecutor)

type defaultTicker struct {
	*time.Ticker
}

func (d *defaultTicker) C() <-chan time.Time {
	return d.Ticker.C
}

type scheduledExecutor struct {
	period       time.Duration
	cb           func()
	isRunning    AtomicBool
	ticker       CustomTicker
	done         chan bool
	wg           sync.WaitGroup
	initialDelay time.Duration
}

func NewScheduledExecutor(period time.Duration, cb func()) ScheduledExecutor {
	return &scheduledExecutor{
		period:       period,
		cb:           cb,
		isRunning:    NewAtomicBool(false),
		wg:           sync.WaitGroup{},
		initialDelay: period,
	}
}

func NewScheduledExecutorWithOpts(period time.Duration, cb func(), options ...Option) ScheduledExecutor {
	s := &scheduledExecutor{
		period:       period,
		cb:           cb,
		isRunning:    NewAtomicBool(false),
		wg:           sync.WaitGroup{},
		initialDelay: period,
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

	s.ticker = CustomTicker(&defaultTicker{time.NewTicker(s.period)})
	s.done = make(chan bool)
	s.wg.Add(1)

	go s.poll()
}

func (s *scheduledExecutor) Stop() {
	if !s.isRunning.Set(false) {
		return
	}

	s.ticker.Stop()
	s.done <- true
	s.wg.Wait()

	s.ticker = nil
	s.done = nil
}

func (s *scheduledExecutor) poll() {
	defer s.wg.Done()
	time.Sleep(s.initialDelay)
	for ; true; <-s.ticker.C() {
		select {
		case <-s.done:
			return
		default:
		}
		s.cb()
	}
}
