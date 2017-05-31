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

func TestStringConversion(t *testing.T) {
	validate := func(t *testing.T, expected string, toTest interface{}) {
		res, ok := asString(toTest)
		assert.True(t, ok)
		assert.Equal(t, expected, res)
	}

	validate(t, "12", int(12))
	validate(t, "13", int32(13))
	validate(t, "14", int64(14))
	validate(t, "12.000000", float32(12.0))
	validate(t, "14.000000", float64(14.0))
	validate(t, "true", true)

	_, ok := asString([]string{"this", "should", "fail"})
	assert.False(t, ok)

	_, ok = asString(map[string]string{"this": "fails"})
	assert.False(t, ok)
}
