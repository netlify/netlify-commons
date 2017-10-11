package transport

import (
	"os"
	"testing"

	"github.com/netlify/netlify-commons/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteToSFX(t *testing.T) {
	token := sfxKey(t)

	trans, err := NewSignalFXTransport(&SFXConfig{token, 0})
	require.NoError(t, err)
	env := metrics.NewEnvironment(trans)
	env.ErrorHandler = func(_ *metrics.RawMetric, err error) {
		assert.Fail(t, "unexpected error: "+err.Error())
	}

	c := env.NewCounter("test.unittest.1", metrics.DimMap{"test": true})
	c.Count(nil)
}

func TestUnsupportedType(t *testing.T) {
	token := sfxKey(t)

	trans, err := NewSignalFXTransport(&SFXConfig{token, 0})
	require.NoError(t, err)
	env := metrics.NewEnvironment(trans)
	env.ErrorHandler = func(_ *metrics.RawMetric, err error) {
		assert.Fail(t, "unexpected error: "+err.Error())
	}

	c := env.NewCounter("test.unittest.2", metrics.DimMap{"test": []bool{true}})
	c.Count(nil)
}

func sfxKey(t *testing.T) string {
	token := os.Getenv("NF_SFX_TOKEN")
	if token == "" {
		t.Skip("NF_SFX_TOKEN not set - skipping tests")
		return ""
	}

	return token
}
