package util

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestDuration_Unmarshal(t *testing.T) {
	testCases := []struct {
		name      string
		url       string
		expected  time.Duration
		errCheck  require.ErrorAssertionFunc
		unmarshal func([]byte, interface{}) error
	}{
		{"empty-json", `{"d": ""}`, 0, require.Error, json.Unmarshal},
		{"invalid-json", `{"d": "no duration here"}`, 0, require.Error, json.Unmarshal},
		{"valid-json-string", `{"d": "1s"}`, time.Second, require.NoError, json.Unmarshal},
		{"valid-json-int", `{"d": 1000000000}`, time.Second, require.NoError, json.Unmarshal},
		{"valid-json-float", `{"d": 1000000000.0}`, time.Second, require.NoError, json.Unmarshal},

		{"empty-yaml", `u: ""}`, 0, require.Error, yaml.Unmarshal},
		{"invalid-yaml", `u: "no duration here"}`, 0, require.Error, yaml.Unmarshal},
		{"valid-yaml-string", `{"d": "1s"}`, time.Second, require.NoError, yaml.Unmarshal},
		{"valid-yaml-int", `{"d": 1000000000}`, time.Second, require.NoError, yaml.Unmarshal},
		{"valid-yaml-float", `{"d": 1000000000.0}`, time.Second, require.NoError, yaml.Unmarshal},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var s struct {
				D Duration `json:"d"`
			}
			tc.errCheck(t, tc.unmarshal([]byte(tc.url), &s))
			assert.Equal(t, tc.expected, s.D.Duration)
		})
	}
}

func TestDuration_Marshal(t *testing.T) {
	testCases := []struct {
		name     string
		marshal  func(interface{}) ([]byte, error)
		expected string
	}{
		{"json", json.Marshal, `{"d":"1s"}`},
		{"yaml", yaml.Marshal, "d: 1s\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := struct {
				D Duration `json:"d" yaml:"d"`
			}{D: Duration{time.Second}}

			serialized, err := tc.marshal(u)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, string(serialized))
		})
	}
}
