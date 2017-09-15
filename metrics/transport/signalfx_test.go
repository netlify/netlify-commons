package transport

import (
	"os"
	"testing"

	"github.com/netlify/netlify-commons/metrics"
	"github.com/stretchr/testify/require"
)

func TestWriteToSFX(t *testing.T) {
	token := sfxKey(t)

	trans, err := NewSignalFXTransport(&SFXConfig{token, 0})
	require.NoError(t, err)
	env := metrics.NewEnvironment(trans)
	c := env.NewCounter("test.unittest.1", metrics.DimMap{"test": true})

	require.NoError(t, c.Count(nil))
}

func TestUnsupportedType(t *testing.T) {
	token := sfxKey(t)

	trans, err := NewSignalFXTransport(&SFXConfig{token, 0})
	require.NoError(t, err)
	env := metrics.NewEnvironment(trans)
	c := env.NewCounter("test.unittest.2", metrics.DimMap{"test": []bool{true}})

	require.Error(t, c.Count(nil))
}

func sfxKey(t *testing.T) string {
	token := os.Getenv("NF_SFX_TOKEN")
	if token == "" {
		t.Skip("NF_SFX_TOKEN not set - skipping tests")
		return ""
	}

	return token
}
