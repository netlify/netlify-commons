package transport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/netlify/netlify-commons/metrics"
	"github.com/pkg/errors"
)

const (
	DataDogIngestEndpoint = "https://app.datadoghq.com/api/v1/series"
	DataDogAuthEndpoint   = "https://app.datadoghq.com/api/v1/validate"
)

var defaultClient = &http.Client{}

type DataDogTransport struct {
	client *http.Client
	apiKey string
	appKey string
}

type DataDogMetric struct {
	Metric string            `json:"metric"`
	Points []DataDogPoint    `json:"points"`
	Tags   map[string]string `json:"tags"`
}

type DataDogPoint []int64

func (t *DataDogPoint) SetTime(when time.Time) {
	(*t)[0] = int64(when.Unix())
}
func (t *DataDogPoint) SetValueInt64(v int64) {
	(*t)[1] = v
}

func NewDataDogPoint(nanos int64, val int64) DataDogPoint {
	when := time.Unix(0, nanos).Unix()
	if nanos == 0 {
		when = time.Now().Unix()
	}
	return DataDogPoint{when, val}
}

func NewDataDogTransport(apiKey, appKey string) (*DataDogTransport, error) {
	return NewDataDogTransportWithClient(apiKey, appKey, defaultClient)
}

func NewDataDogTransportWithClient(apiKey, appKey string, client *http.Client) (*DataDogTransport, error) {
	trans := &DataDogTransport{
		client: client,
		apiKey: apiKey,
		appKey: appKey,
	}

	if err := trans.validate(); err != nil {
		return nil, err
	}

	return trans, nil
}

func (t DataDogTransport) Publish(m *metrics.RawMetric) error {
	point := DataDogMetric{
		Metric: m.Name,
		Points: []DataDogPoint{},
		Tags:   map[string]string{},
	}

	for k, v := range m.Dims {
		if val, ok := asString(v); ok {
			point.Tags[k] = val
		}
	}

	point.Points = append(point.Points, NewDataDogPoint(m.Timestamp, m.Value))

	msg := &struct {
		Series []DataDogMetric `json:"series"`
	}{
		Series: []DataDogMetric{point},
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Failed to serialize %v", msg))
	}

	req, err := http.NewRequest(http.MethodPost, DataDogIngestEndpoint, bytes.NewReader(b))
	if err != nil {
		return errors.Wrap(err, "Creating request failed")
	}

	return t.do(req, http.StatusAccepted)
}

func (t DataDogTransport) validate() error {
	req, err := http.NewRequest(http.MethodGet, DataDogAuthEndpoint, nil)
	if err != nil {
		return errors.Wrap(err, "Creating request failed")
	}

	return t.do(req, http.StatusOK)
}

func (t DataDogTransport) do(req *http.Request, expectedStatus int) error {
	q := req.URL.Query()
	q.Add("api_key", t.apiKey)
	q.Add("app_key", t.appKey)
	req.URL.RawQuery = q.Encode()

	rsp, err := t.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "Failed to make authentication request")
	}

	if rsp.StatusCode != expectedStatus {
		defer rsp.Body.Close()
		body, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return fmt.Errorf("Failed to get a 200 from Datadog. Got %d.", rsp.StatusCode)
		}
		return fmt.Errorf("Failed to get a 200 from Datadog. Got %d and body: %s", rsp.StatusCode, string(body))
	}

	return nil
}

func asString(v interface{}) (string, bool) {
	switch v.(type) {
	case string:
		return v.(string), true
	case int, int64, int32:
		return fmt.Sprintf("%d", v), true
	case float32, float64:
		return fmt.Sprintf("%f", v), true
	case bool:
		return fmt.Sprintf("%t", v), true
	}

	return "", false
}
