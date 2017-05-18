package metrics

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func NewEnvironment(trans Transport) *Environment {
	return &Environment{
		dimlock:    sync.Mutex{},
		globalDims: DimMap{},
		transport:  trans,
	}
}

type Environment struct {
	globalDims DimMap
	dimlock    sync.Mutex

	tracer    func(m *RawMetric)
	transport Transport

	// some metrics stuff
	timersSent   int64
	countersSent int64
	gaugesSent   int64
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
	default:
		return UnknownMetricTypeError{errString{fmt.Sprintf("unknown metric type: %s", m.Type)}}
	}

	if e.tracer != nil {
		e.tracer(m)
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
