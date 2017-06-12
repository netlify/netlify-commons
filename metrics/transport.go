package metrics

type Transport interface {
	Publish(m *RawMetric) error
	PublishEvent(e *Event) error
}

type NoopTransport struct{}

func (NoopTransport) Publish(_ *RawMetric) error {
	return nil
}

func (NoopTransport) PublishEvent(_ *Event) error {
	return nil
}

func TransportFunc(mets func(m *RawMetric) error, events func(e *Event) error) Transport {
	return transportWrapper{mets, events}
}

type transportWrapper struct {
	metrics func(*RawMetric) error
	events  func(*Event) error
}

func (t transportWrapper) Publish(m *RawMetric) error {
	if t.metrics != nil {
		return t.metrics(m)
	}
	return nil
}

func (t transportWrapper) PublishEvent(e *Event) error {
	if t.events != nil {
		return t.events(e)
	}
	return nil
}

func NewBroadcastTransport(ports []Transport) *BroadcastTransport {
	return &BroadcastTransport{ports}
}

type BroadcastTransport struct {
	ports []Transport
}

func (t BroadcastTransport) Publish(m *RawMetric) error {
	for _, p := range t.ports {
		if err := p.Publish(m); err != nil {
			return err
		}
	}
	return nil
}

func (t BroadcastTransport) PublishEvent(e *Event) error {
	for _, p := range t.ports {
		if err := p.PublishEvent(e); err != nil {
			return err
		}
	}
	return nil
}
