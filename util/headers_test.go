package util

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestHeaders_Unmarshal(t *testing.T) {
	testCases := []struct {
		name      string
		url       string
		expected  http.Header
		errCheck  require.ErrorAssertionFunc
		unmarshal func([]byte, interface{}) error
	}{
		{"invalid-json", `{"h": "<>:hey>"}`, nil, require.Error, json.Unmarshal},
		{"null-json", `{"h": null}`, http.Header{}, require.NoError, json.Unmarshal},
		{"empty-json", `{"h": {}}`, http.Header{}, require.NoError, json.Unmarshal},
		{
			"valid-json",
			`{"h": {"X-NF-TEST-A": "aAa", "x-nf-test-b": "bBb"}}`,
			http.Header{"X-Nf-Test-A": {"aAa"}, "X-Nf-Test-B": {"bBb"}},
			require.NoError,
			json.Unmarshal,
		},
		{
			"duplicate-json",
			`{"h": {"X-NF-TEST-A": "aAa", "X-Nf-Test-A": "bBb", "x-nf-test-a": "cCc"}}`,
			http.Header{"X-Nf-Test-A": {"aAa", "bBb", "cCc"}},
			require.NoError,
			json.Unmarshal,
		},

		{"invalid-yaml", `h: ""`, nil, require.Error, yaml.Unmarshal},
		{"null-yaml", `h: null`, http.Header{}, require.NoError, yaml.Unmarshal},
		{"empty-yaml", `h: {}`, http.Header{}, require.NoError, yaml.Unmarshal},
		{
			"valid-yaml",
			`h: {"X-NF-TEST-A": "aAa", "x-nf-test-b": "bBb"}`,
			http.Header{"X-Nf-Test-A": {"aAa"}, "X-Nf-Test-B": {"bBb"}},
			require.NoError,
			yaml.Unmarshal,
		},
		{
			"duplicate-yaml",
			`h: {"X-NF-TEST-A": "aAa", "X-Nf-Test-A": "bBb", "x-nf-test-a": "cCc"}`,
			http.Header{"X-Nf-Test-A": {"aAa", "bBb", "cCc"}},
			require.NoError,
			yaml.Unmarshal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var s struct {
				H Headers `json:"h"`
			}
			tc.errCheck(t, tc.unmarshal([]byte(tc.url), &s))
			assert.ObjectsAreEqualValues(tc.expected, s.H.Header)
		})
	}
}

func TestHeaders_Marshal(t *testing.T) {
	testCases := []struct {
		name     string
		marshal  func(interface{}) ([]byte, error)
		expected string
	}{
		{"json", json.Marshal, `{"h":{"X-Nf-Test-A":["aAa","bBb"]}}`},
		{"yaml", yaml.Marshal, "h:\n    X-Nf-Test-A:\n        - aAa\n        - bBb\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := struct {
				H Headers `json:"h" yaml:"h"`
			}{H: Headers{http.Header{"X-Nf-Test-A": {"aAa", "bBb"}}}}

			serialized, err := tc.marshal(h)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, string(serialized))
		})
	}
}
