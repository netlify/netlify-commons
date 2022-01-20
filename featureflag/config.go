package featureflag

import (
	"github.com/netlify/netlify-commons/util"
	"gopkg.in/launchdarkly/go-server-sdk.v5/interfaces"
)

type Config struct {
	Key            string        `json:"key" yaml:"key"`
	RequestTimeout util.Duration `json:"request_timeout" yaml:"request_timeout" mapstructure:"request_timeout" split_words:"true" default:"5s"`
	Enabled        bool          `json:"enabled" yaml:"enabled" default:"false"`

	updateProcessorFactory interfaces.DataSourceFactory

	// Drop telemetry events (not needed in local-dev/CI environments)
	DisableEvents bool `json:"disable_events" yaml:"disable_events" mapstructure:"disable_events" split_words:"true"`

	// Set when using the Launch Darkly Relay proxy
	RelayHost string `json:"relay_host" yaml:"relay_host" mapstructure:"relay_host" split_words:"true"`

	// DefaultUserAttrs are custom LaunchDarkly user attributes that are added to every
	// feature flag check
	DefaultUserAttrs map[string]string `json:"default_user_attrs" yaml:"default_user_attrs"`
}
