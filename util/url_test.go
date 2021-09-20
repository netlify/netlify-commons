package util

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestURL_Unmarshal(t *testing.T) {
	testCases := []struct {
		name      string
		url       string
		expected  *url.URL
		errCheck  require.ErrorAssertionFunc
		unmarshal func([]byte, interface{}) error
	}{
		{"empty-json", `{"u": ""}`, nil, require.Error, json.Unmarshal},
		{"invalid-json", `{"u": "<>:hey>"}`, nil, require.Error, json.Unmarshal},
		{"valid-json", `{"u": "https://netlify.com"}`, &url.URL{Scheme: "https", Host: "netlify.com"}, require.NoError, json.Unmarshal},

		{"empty-yaml", `u: ""}`, nil, require.Error, yaml.Unmarshal},
		{"invalid-yaml", `u: "<>:hey>"}`, nil, require.Error, yaml.Unmarshal},
		{"valid-yaml", `{u: https://netlify.com}`, &url.URL{Scheme: "https", Host: "netlify.com"}, require.NoError, yaml.Unmarshal},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var s struct {
				U URL `json:"u"`
			}
			tc.errCheck(t, tc.unmarshal([]byte(tc.url), &s))
			assert.Equal(t, tc.expected, s.U.URL)
		})
	}
}

func TestURL_Marshal(t *testing.T) {
	testCases := []struct {
		name     string
		marshal  func(interface{}) ([]byte, error)
		expected string
	}{
		{"json", json.Marshal, `{"u":"https://netlify.com"}`},
		{"yaml", yaml.Marshal, "u: https://netlify.com\n"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := struct {
				U URL `json:"u" yaml:"u"`
			}{U: URL{&url.URL{Scheme: "https", Host: "netlify.com"}}}

			serialized, err := tc.marshal(u)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, string(serialized))
		})
	}
}
