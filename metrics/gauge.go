package metrics

import (
	"sync"
	"time"
)

// Gauge keeps a running measure of the value at that moment
type Gauge interface {
	Increment(DimMap)
	Decrement(DimMap)
	Set(int64, DimMap)
	SetTimestamp(time.Time)
}

type gauge struct {
	metric
	valueLock *sync.Mutex
}

func (e *Environment) NewGauge(name string, metricDims DimMap) Gauge {
	m := e.newMetric(name, GaugeType, metricDims)
	return &gauge{
		metric:    *m,
		valueLock: new(sync.Mutex),
	}
}

func (m *gauge) Increment(instanceDims DimMap) {
	m.valueLock.Lock()
	defer m.valueLock.Unlock()
	m.Value++
	m.send(instanceDims)
}

func (m *gauge) Decrement(instanceDims DimMap) {
	m.valueLock.Lock()
	defer m.valueLock.Unlock()
	m.Value--
	m.send(instanceDims)
}

func (m *gauge) Set(val int64, instanceDims DimMap) {
	m.valueLock.Lock()
	defer m.valueLock.Unlock()
	m.Value = val
	m.send(instanceDims)
}
