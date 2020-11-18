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

	// Drop telemetry events (not needed in local-dev/CI environments)
	DisableEvents bool `mapstructure:"disable_events" split_words:"true"`

	// Set when using the Launch Darkly Relay proxy
	RelayHost string `mapstructure:"relay_host" split_words:"true"`
}
