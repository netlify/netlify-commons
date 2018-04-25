package transport

import (
	"encoding/json"

	"github.com/nats-io/go-nats"
	"github.com/pkg/errors"

	"github.com/netlify/netlify-commons/metrics"
)

func NewNatsTransport(subject string, conn *nats.Conn) *NatsTransport {
	return &NatsTransport{subject, conn}
}

type NatsTransport struct {
	subject string
	conn    *nats.Conn
}

func (t *NatsTransport) Publish(m *metrics.RawMetric) error {
	if t.subject == "" {
		return errors.New("No subject provided.")
	}

	data, err := json.Marshal(m)
	if err != nil {
		return errors.Wrap(err, "marshalling to json failed")
	}

	return t.conn.Publish(t.subject, data)
}

func (t *NatsTransport) Queue(m *metrics.RawMetric) error {
	// NATS does not support queueing
	return t.Publish(m)
}
