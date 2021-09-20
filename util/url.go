package util

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/pkg/errors"
)

// URL is a serializable version version of a url.URL
// it supports serialization in yaml and json
type URL struct {
	*url.URL
}

func (u URL) MarshalYAML() (interface{}, error) {
	return u.String(), nil
}

func (u URL) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.String())
}

func (u *URL) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v interface{}
	if err := unmarshal(&v); err != nil {
		return err
	}
	return u.setValue(v)
}

func (u *URL) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	return u.setValue(v)
}

func (u *URL) setValue(v interface{}) error {
	switch value := v.(type) {
	case string:
		if value == "" {
			return errors.New("empty string provided as url")
		}

		parsed, err := url.Parse(value)
		if err != nil {
			return errors.Wrap(err, "invalid url provided")
		}

		u.URL = parsed
	default:
		return errors.New(fmt.Sprintf("invalid type provided as url: %T", v))
	}
	return nil
}
