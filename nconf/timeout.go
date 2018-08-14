package nconf

import "time"

// HTTPServerTimeoutConfig represents common HTTP server timeout values
type HTTPServerTimeoutConfig struct {
	// Read = http.Server.ReadTimeout
	Read time.Duration `mapstructure:"read"`
	// Write = http.Server.WriteTimeout
	Write time.Duration `mapstructure:"write"`
	// Handler = http.TimeoutHandler (or equivalent).
	// The maximum amount of time a server handler can take.
	Handler time.Duration `mapstructure:"handler"`
}

// HTTPClientTimeoutConfig represents common HTTP client timeout values
type HTTPClientTimeoutConfig struct {
	// Dial = net.Dialer.Timeout
	Dial time.Duration `mapstructure:"dial"`
	// KeepAlive = net.Dialer.KeepAlive
	KeepAlive time.Duration `mapstructure:"keep_alive" split_words:"true"`

	// TLSHandshake = http.Transport.TLSHandshakeTimeout
	TLSHandshake time.Duration `mapstructure:"tls_handshake" split_words:"true"`
	// ResponseHeader = http.Transport.ResponseHeaderTimeout
	ResponseHeader time.Duration `mapstructure:"response_header" split_words:"true"`
	// Total = http.Client.Timeout or equivalent
	// The maximum amount of time a client request can take.
	Total time.Duration `mapstructure:"total"`
}
