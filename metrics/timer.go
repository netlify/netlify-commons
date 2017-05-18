package metrics

import "time"

// Timer will measure the time between two events and send that
type Timer interface {
	Start() time.Time
	Stop(instanceDims DimMap) (time.Duration, error)
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

func (t *timer) Stop(instanceDims DimMap) (time.Duration, error) {
	now := time.Now()

	if t.startTime == nil {
		return 0, NotStartedError{errString{"the timer hasn't been started yet"}}
	}

	diff := now.Sub(*t.startTime)
	t.Value = int64(diff)

	return diff, t.send(instanceDims)
}

func (e *Environment) timeBlock(name string, metricDims DimMap, f func()) time.Duration {
	t := e.NewTimer(name, metricDims)
	t.Start()
	f()
	dur, _ := t.Stop(nil)
	return dur
}

func (e *Environment) timeBlockErr(name string, metricDims DimMap, f func() error) (time.Duration, error) {
	t := e.NewTimer(name, metricDims)
	t.Start()
	err := f()
	dur, _ := t.Stop(DimMap{"had_error": err != nil})

	return dur, err
}
