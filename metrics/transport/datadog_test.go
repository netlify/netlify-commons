package transport

import (
	"os"
	"testing"

	"github.com/netlify/netlify-commons/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendToDataDog(t *testing.T) {
	apikey, appkey := datadogKey(t)

	trans, err := NewDataDogTransport(apikey, appkey)
	require.NoError(t, err)

	m := &metrics.RawMetric{
		Name:  "testmetric",
		Value: 123,
		Dims: metrics.DimMap{
			"test":   "value",
			"ignore": true,
		},
	}
	assert.NoError(t, trans.Publish(m))
}

func datadogKey(t *testing.T) (string, string) {

	apikey := os.Getenv("NF_DD_API_KEY")
	if apikey == "" {
		t.Skip("NF_DD_API_KEY not set - skipping tests")
	}

	appkey := os.Getenv("NF_DD_APP_KEY")
	if appkey == "" {
		t.Skip("NF_DD_APP_KEY not set - skipping tests")
	}

	return apikey, appkey
}
