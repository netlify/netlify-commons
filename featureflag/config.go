package featureflag

import "time"

type Config struct {
	Key            string
	RequestTimeout time.Duration `mapstructure:"request_timeout" split_words:"true" default:"5s"`
}
