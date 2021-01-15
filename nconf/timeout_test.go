package nconf

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestParseTimeoutValues(t *testing.T) {
	raw := map[string]string{
		"dial":            "10s",
		"keep_alive":      "11s",
		"tls_handshake":   "12s",
		"response_header": "13s",
		"total":           "14s",
	}

	// write it to json & yaml
	// then load it through the RootArgs.load
	scenes := []struct {
		name string
		enc  func(interface{}) ([]byte, error)
		dec  func([]byte, interface{}) error
	}{
		{"json", json.Marshal, json.Unmarshal},
		{"yaml", yaml.Marshal, yaml.Unmarshal},
	}
	for _, s := range scenes {
		t.Run(s.name, func(t *testing.T) {
			bs, err := s.enc(&raw)
			require.NoError(t, err)

			var cfg HTTPClientTimeoutConfig

			require.NoError(t, s.dec(bs, &cfg))

			assert.Equal(t, "10s", cfg.Dial.String())
			assert.Equal(t, "11s", cfg.KeepAlive.String())
			assert.Equal(t, "12s", cfg.TLSHandshake.String())
			assert.Equal(t, "13s", cfg.ResponseHeader.String())
			assert.Equal(t, "14s", cfg.Total.String())
			assert.Equal(t, 10*time.Second, cfg.Dial.Duration)
			assert.Equal(t, 11*time.Second, cfg.KeepAlive.Duration)
			assert.Equal(t, 12*time.Second, cfg.TLSHandshake.Duration)
			assert.Equal(t, 13*time.Second, cfg.ResponseHeader.Duration)
			assert.Equal(t, 14*time.Second, cfg.Total.Duration)
		})
	}
}
