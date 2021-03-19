package util

import (
	"sync"
	"time"
)

type ScheduledExecutor interface {
	Start()
	Stop()
}

type scheduledExecutor struct {
	period             time.Duration
	cb                 func()
	isRunning          AtomicBool
	ticker             *time.Ticker
	done               chan bool
	wg                 sync.WaitGroup
	enableInstantStart bool
}

func NewScheduledExecutor(period time.Duration, cb func()) ScheduledExecutor {
	return &scheduledExecutor{
		period:    period,
		cb:        cb,
		isRunning: NewAtomicBool(false),
		wg:        sync.WaitGroup{},
	}
}

func NewExecutorWithInstantStart(period time.Duration, cb func()) ScheduledExecutor {
	return &scheduledExecutor{
		period:             period,
		cb:                 cb,
		isRunning:          NewAtomicBool(false),
		wg:                 sync.WaitGroup{},
		enableInstantStart: true,
	}
}

func (s *scheduledExecutor) Start() {
	if s.isRunning.Set(true) {
		return
	}

	s.ticker = time.NewTicker(s.period)
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
	if s.enableInstantStart {
		for ; true; <-s.ticker.C {
			select {
			case <-s.done:
				return
			default:
			}
			s.cb()
		}
	} else {
		for {
			select {
			case <-s.done:
				return
			case <-s.ticker.C:
				s.cb()
			}
		}
	}
}
