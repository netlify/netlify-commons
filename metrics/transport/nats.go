package transport

import (
	"encoding/json"

	"github.com/nats-io/nats"
	"github.com/pkg/errors"

	"github.com/netlify/netlify-commons/metrics"
)

func NewNatsTransport(metricsSubject, eventsSubject string, conn *nats.Conn) *NatsTransport {
	return &NatsTransport{
		MetricsSubject: metricsSubject,
		EventsSubject:  eventsSubject,
		conn:           conn,
	}
}

type NatsTransport struct {
	MetricsSubject string
	EventsSubject  string
	conn           *nats.Conn
}

func (t *NatsTransport) Publish(m *metrics.RawMetric) error {
	if t.MetricsSubject == "" {
		return errors.New("No subject provided.")
	}

	data, err := json.Marshal(m)
	if err != nil {
		return errors.Wrap(err, "marshalling to json failed")
	}

	return t.conn.Publish(t.MetricsSubject, data)
}

func (t *NatsTransport) PublishEvent(e *metrics.Event) error {
	if t.EventsSubject == "" {
		return errors.New("No subject provided.")
	}

	data, err := json.Marshal(e)
	if err != nil {
		return errors.Wrap(err, "marshalling to json failed")
	}

	return t.conn.Publish(t.EventsSubject, data)
}
