package metrics

import (
	"sync"
	"time"
)

// Gauge keeps a running measure of the value at that moment
type Gauge interface {
	Increment(DimMap) error
	Decrement(DimMap) error
	Set(int64, DimMap) error
	SetTimestamp(time.Time)
}

type gauge struct {
	metric
	valueLock sync.Mutex
}

func (e *Environment) NewGauge(name string, metricDims DimMap) Gauge {
	m := e.newMetric(name, GaugeType, metricDims)
	return &gauge{
		metric:    *m,
		valueLock: sync.Mutex{},
	}
}

func (m *gauge) Increment(instanceDims DimMap) error {
	m.valueLock.Lock()
	defer m.valueLock.Unlock()
	m.Value++
	return m.send(instanceDims)
}

func (m *gauge) Decrement(instanceDims DimMap) error {
	m.valueLock.Lock()
	defer m.valueLock.Unlock()
	m.Value--
	return m.send(instanceDims)
}

func (m *gauge) Set(val int64, instanceDims DimMap) error {
	m.valueLock.Lock()
	defer m.valueLock.Unlock()
	m.Value = val
	return m.send(instanceDims)
}
