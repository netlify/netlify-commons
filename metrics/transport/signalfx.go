package transport

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/pkg/errors"
	"github.com/signalfx/golib/datapoint"
	"github.com/signalfx/golib/event"
	"github.com/signalfx/golib/sfxclient"

	"github.com/netlify/netlify-commons/metrics"
)

type SFXConfig struct {
	AuthToken string `mapstructure:"auth_token"`
}

func NewSignalFXTransport(config *SFXConfig) (*SFXTransport, error) {
	sink := sfxclient.NewHTTPSink()
	sink.AuthToken = config.AuthToken

	return &SFXTransport{sink}, nil
}

type SFXTransport struct {
	sink *sfxclient.HTTPSink
}

func (t *SFXTransport) Publish(m *metrics.RawMetric) error {
	p := &datapoint.Datapoint{
		Metric:    m.Name,
		Value:     datapoint.NewIntValue(m.Value),
		Timestamp: time.Unix(0, m.Timestamp),
	}
	converted, err := convert(m.Dims)
	if err != nil {
		return err
	}

	p.Dimensions = converted

	return t.sink.AddDatapoints(context.Background(), []*datapoint.Datapoint{p})
}

func (t *SFXTransport) PublishEvent(e *metrics.Event) error {
	dims, err := convert(e.Dims)
	if err != nil {
		return errors.Wrap(err, "Failed to convert dims to strings")
	}

	sfxEvent := &event.Event{
		EventType:  e.Name,
		Timestamp:  time.Unix(0, e.Timestamp),
		Dimensions: dims,
		Properties: e.Props,
	}

	return t.sink.AddEvents(context.Background(), []*event.Event{sfxEvent})
}

func convert(in metrics.DimMap) (map[string]string, error) {
	res := map[string]string{}
	for k, v := range in {
		var asStr string
		switch v.(type) {
		case int, int64, int32:
			asStr = fmt.Sprintf("%d", v)
		case bool:
			asStr = fmt.Sprintf("%t", v)
		case float32, float64:
			asStr = fmt.Sprintf("%f", v)
		case string:
			asStr = v.(string)
		default:
			return nil, fmt.Errorf("Unsupported type for dimension '%s': '%v'", k, reflect.TypeOf(v))
		}
		res[k] = asStr
	}

	return res, nil

}
