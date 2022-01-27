package util

import (
	"encoding/json"
	"net/http"
)

// Headers is a serializable version of http.Header it supports both yaml & json formats.
// Headers expects the yaml/json representation to be a map[string]string.
type Headers struct {
	http.Header
}

func (h Headers) MarshalYAML() (interface{}, error) {
	return h.Header, nil
}

func (h Headers) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.Header)
}

func (h *Headers) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var headers map[string]string
	if err := unmarshal(&headers); err != nil {
		return err
	}

	h.Header = http.Header{}
	for k, v := range headers {
		h.Add(k, v)
	}

	return nil
}

func (h *Headers) UnmarshalJSON(b []byte) error {
	var headers map[string]string
	if err := json.Unmarshal(b, &headers); err != nil {
		return err
	}

	h.Header = http.Header{}
	for k, v := range headers {
		h.Add(k, v)
	}

	return nil
}
