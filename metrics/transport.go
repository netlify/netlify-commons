package metrics

type Transport interface {
	Publish(m *RawMetric) error
}

type NoopTransport struct{}

func (NoopTransport) Publish(_ *RawMetric) error {
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

func NewBroadcastTransport(ports []Transport) Transport {
	return TransportFunc(func(m *RawMetric) error {
		for _, p := range ports {
			if err := p.Publish(m); err != nil {
				return err
			}
		}
		return nil
	})
}
