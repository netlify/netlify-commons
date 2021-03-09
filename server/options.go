package server

import (
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
