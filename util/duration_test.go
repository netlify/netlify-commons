package util

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestDurationParsing(t *testing.T) {
	scenes := []struct {
		name string
		dur  interface{}
	}{
		{"float", 1e9},
		{"int", int(1e9)},
		{"str", "1s"},
	}

	for _, s := range scenes {
		t.Run(s.name, func(t *testing.T) {
			cfg := struct {
				Dur interface{}
			}{
				Dur: s.dur,
			}
			t.Run("yaml", func(t *testing.T) {
				bs, err := yaml.Marshal(&cfg)
				require.NoError(t, err)

				res := struct {
					Dur Duration
				}{}
				require.NoError(t, yaml.Unmarshal(bs, &res))
				assert.Equal(t, time.Second, res.Dur.Duration)
			})

			t.Run("json", func(t *testing.T) {
				bs, err := json.Marshal(&cfg)
				require.NoError(t, err)
				res := struct {
					Dur Duration
				}{}

				require.NoError(t, json.Unmarshal(bs, &res))
				assert.Equal(t, time.Second, res.Dur.Duration)
			})
		})
	}
}
