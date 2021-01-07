package nconf

import (
	"github.com/netlify/netlify-commons/util"
)

// HTTPServerTimeoutConfig represents common HTTP server timeout values
type HTTPServerTimeoutConfig struct {
	// Read = http.Server.ReadTimeout
	Read util.Duration `mapstructure:"read"`
	// Write = http.Server.WriteTimeout
	Write util.Duration `mapstructure:"write"`
	// Handler = http.TimeoutHandler (or equivalent).
	// The maximum amount of time a server handler can take.
	Handler util.Duration `mapstructure:"handler"`
}

// HTTPClientTimeoutConfig represents common HTTP client timeout values
type HTTPClientTimeoutConfig struct {
	// Dial = net.Dialer.Timeout
	Dial util.Duration `mapstructure:"dial"`
	// KeepAlive = net.Dialer.KeepAlive
	KeepAlive util.Duration `mapstructure:"keep_alive" split_words:"true" json:"keep_alive" yaml:"keep_alive"`

	// TLSHandshake = http.Transport.TLSHandshakeTimeout
	TLSHandshake util.Duration `mapstructure:"tls_handshake" split_words:"true" json:"tls_handshake" yaml:"tls_handshake"`
	// ResponseHeader = http.Transport.ResponseHeaderTimeout
	ResponseHeader util.Duration `mapstructure:"response_header" split_words:"true" json:"response_header" yaml:"response_header"`
	// Total = http.Client.Timeout or equivalent
	// The maximum amount of time a client request can take.
	Total util.Duration `mapstructure:"total"`
}
