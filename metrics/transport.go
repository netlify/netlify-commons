package metrics

type Transport interface {
	Publish(m *RawMetric) error
	Queue(m *RawMetric) error
}

type NoopTransport struct{}

func (NoopTransport) Publish(_ *RawMetric) error {
	return nil
}

func (NoopTransport) Queue(_ *RawMetric) error {
	return nil
}

func TransportFunc(f func(m *RawMetric) error) Transport {
	return transportWrapper{f}
}

type transportWrapper struct {
	f func(*RawMetric) error
}

func (t transportWrapper) Publish(m *RawMetric) error {
	return t.f(m)
}

func (t transportWrapper) Queue(m *RawMetric) error {
	return t.f(m)
}
