package util

import (
	"encoding/json"
	"errors"
	"time"

	"gopkg.in/yaml.v3"
)

// Duration is a serializable version version of a time.Duration
// it supports setting in yaml & json via:
// - string: 10s
// - float32/64, int/32/64: 10 (nanoseconds)
type Duration struct {
	time.Duration
}

func (d Duration) MarshalYAML() ([]byte, error) {
	return yaml.Marshal(d.String())
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v interface{}
	if err := unmarshal(&v); err != nil {
		return err
	}
	return d.setValue(v)
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	return d.setValue(v)
}

func (d *Duration) UnmarshalText(text []byte) error {
	return d.setValue(string(text))
}

func (d *Duration) setValue(v interface{}) error {
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
	case float32:
		d.Duration = time.Duration(value)
	case int:
		d.Duration = time.Duration(value)
	case int32:
		d.Duration = time.Duration(value)
	case int64:
		d.Duration = time.Duration(value)
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid duration")
	}
	return nil
}
