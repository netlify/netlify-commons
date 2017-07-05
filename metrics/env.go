package metrics

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func NewEnvironment(trans Transport) *Environment {
	return &Environment{
		dimlock:    sync.Mutex{},
		globalDims: DimMap{},
		transport:  trans,
		reporter:   new(noopReporter),
	}
}

type Environment struct {
	globalDims DimMap
	dimlock    sync.Mutex

	Namespace string
	Tracer    func(m *RawMetric)
	transport Transport

	reporter reporter

	// some metrics stuff
	timersSent      int64
	countersSent    int64
	gaugesSent      int64
	cumulativesSent int64
}

func (e *Environment) StartReportingCumulativeCounters(interval time.Duration, log *logrus.Entry) {
	if interval.Seconds() > 0 {
		e.reporter = newIntervalReporter(interval, log)
		e.reporter.start()
	}
}

func (e *Environment) send(m *RawMetric) error {
	if m == nil {
		return nil
	}

	switch m.Type {
	case CounterType:
		atomic.AddInt64(&e.countersSent, 1)
	case TimerType:
		atomic.AddInt64(&e.timersSent, 1)
	case GaugeType:
		atomic.AddInt64(&e.gaugesSent, 1)
	case CumulativeType:
		atomic.AddInt64(&e.cumulativesSent, 1)
	default:
		return UnknownMetricTypeError{errors.Errorf("unknown metric type: %s", m.Type)}
	}

	if e.Tracer != nil {
		e.Tracer(m)
	}

	if e.Namespace != "" {
		m.Name = e.Namespace + m.Name
	}

	return e.transport.Publish(m)
}

func (e *Environment) AddDimension(k string, v interface{}) {
	e.dimlock.Lock()
	defer e.dimlock.Unlock()
	e.globalDims[k] = v
}

func addAll(into DimMap, from DimMap) {
	if into != nil {
		if from != nil {
			for k, v := range from {
				into[k] = v
			}
		}
	}
}

func (e *Environment) newMetric(name string, t MetricType, dims DimMap) *metric {
	m := &metric{
		RawMetric: RawMetric{
			Name: name,
			Type: t,
			Dims: make(DimMap),
		},
		env:     e,
		dimlock: &sync.Mutex{},
	}

	if dims != nil {
		for k, v := range dims {
			m.AddDimension(k, v)
		}
	}
	return m
}
