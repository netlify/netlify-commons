package featureflag

import (
	"github.com/netlify/netlify-commons/util"
	ld "gopkg.in/launchdarkly/go-server-sdk.v4"
)

type Config struct {
	Key            string        `json:"key" yaml:"key"`
	RequestTimeout util.Duration `json:"request_timeout" yaml:"request_timeout" mapstructure:"request_timeout" split_words:"true" default:"5s"`
	Enabled        bool          `json:"enabled" yaml:"enabled" default:"false"`

	updateProcessorFactory ld.UpdateProcessorFactory

	// Drop telemetry events (not needed in local-dev/CI environments)
	DisableEvents bool `json:"disable_events" yaml:"disable_events" mapstructure:"disable_events" split_words:"true"`

	// Set when using the Launch Darkly Relay proxy
	RelayHost string `json:"relay_host" yaml:"relay_host" mapstructure:"relay_host" split_words:"true"`
}
