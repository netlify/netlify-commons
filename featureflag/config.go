package featureflag

import (
	"time"

	ld "gopkg.in/launchdarkly/go-server-sdk.v4"
)

type Config struct {
	Key                    string
	RequestTimeout         time.Duration `mapstructure:"request_timeout" split_words:"true" default:"5s"`
	Enabled                bool          `default:"false"`
	updateProcessorFactory ld.UpdateProcessorFactory
}
