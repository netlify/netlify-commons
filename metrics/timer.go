package metrics

import (
	"time"

	"github.com/pkg/errors"
)

// Timer will measure the time between two events and send that
type Timer interface {
	Start() time.Time
	Stop(instanceDims DimMap) time.Duration
	SetTimestamp(time.Time)
}

type timer struct {
	metric
	startTime *time.Time
}

func (e *Environment) NewTimer(name string, metricDims DimMap) Timer {
	m := e.newMetric(name, TimerType, metricDims)
	return &timer{
		metric: *m,
	}
}

func (t *timer) Start() time.Time {
	now := time.Now()
	t.startTime = &now
	return now
}

func (t *timer) Stop(instanceDims DimMap) time.Duration {
	now := time.Now()

	if t.startTime == nil {
		t.env.reportError(&t.RawMetric, NotStartedError{errors.New("the timer hasn't been started yet")})
	}

	diff := now.Sub(*t.startTime)
	t.Value = int64(diff)
	t.send(instanceDims)
	return diff
}

func (e *Environment) timeBlock(name string, metricDims DimMap, f func()) time.Duration {
	t := e.NewTimer(name, metricDims)
	t.Start()
	f()
	return t.Stop(nil)
}

func (e *Environment) timeBlockErr(name string, metricDims DimMap, f func() error) (time.Duration, error) {
	t := e.NewTimer(name, metricDims)
	t.Start()
	err := f()
	dur := t.Stop(DimMap{"had_error": err != nil})
	return dur, err
}
