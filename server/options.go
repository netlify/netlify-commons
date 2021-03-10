package server

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

// Opt will allow modification of the http server
type Opt func(s *http.Server)

// WithWriteTimeout will override the server's write timeout
func WithWriteTimeout(dur time.Duration) Opt {
	return func(s *http.Server) {
		s.WriteTimeout = dur
	}
}

// WithReadTimeout will override the server's read timeout
func WithReadTimeout(dur time.Duration) Opt {
	return func(s *http.Server) {
		s.ReadTimeout = dur
	}
}

// WithTLS will use the provided TLS configuration
func WithTLS(cfg *tls.Config) Opt {
	return func(s *http.Server) {
		s.TLSConfig = cfg
	}
}

// WithAddress will set the address field on the server
func WithAddress(addr string) Opt {
	return func(s *http.Server) {
		s.Addr = addr
	}
}

// WithHostAndPort will use them in the form host:port as the address field on the server
func WithHostAndPort(host string, port int) Opt {
	return WithAddress(fmt.Sprintf("%s:%d", host, port))
}
