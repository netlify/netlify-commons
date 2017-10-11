package metrics

import "time"

// Counter will send when an event occurs
type Counter interface {
	SetTimestamp(time.Time)
	Count(DimMap)
	CountN(int64, DimMap)
}

func (e *Environment) NewCounter(name string, metricDims DimMap) Counter {
	return e.newMetric(name, CounterType, metricDims)
}

// Count will count 1 occurrence of an event
func (m *metric) Count(instanceDims DimMap) {
	m.CountN(1, instanceDims)
}

//CountN will count N occurrences of an event
func (m *metric) CountN(val int64, instanceDims DimMap) {
	m.Value = val
	m.send(instanceDims)
}
