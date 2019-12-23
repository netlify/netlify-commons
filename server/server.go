package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/netlify/netlify-commons/nconf"
	"github.com/netlify/netlify-commons/router"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Server handles the setup and shutdown of the http server
// for an API
type Server struct {
	log logrus.FieldLogger
	svr *http.Server
}

type Config struct {
	HealthPath string `split_words:"true"`
	Port       int
	TLS        nconf.TLSConfig
}

// APIDefinition is used to define the routes used by the API
type APIDefinition interface {
	AddRoutes(r router.Router)
}

func New(log logrus.FieldLogger, projectName string, config Config, api APIDefinition) (*Server, error) {
	r := router.New(
		log,
		router.OptHealthCheck(config.HealthPath, nil),
		router.OptTracingMiddleware(log, projectName),
	)

	api.AddRoutes(r)

	s := Server{
		log: log.WithField("component", "server"),
		svr: &http.Server{
			Addr:    fmt.Sprintf(":%d", config.Port),
			Handler: r,
		},
	}

	if config.TLS.Enabled {
		tcfg, err := config.TLS.TLSConfig()
		if err != nil {
			return nil, errors.Wrap(err, "Failed to build TLS config")
		}
		s.svr.TLSConfig = tcfg
		log.Info("TLS enabled")
	}

	return &s, nil
}

func (s *Server) Shutdown(to time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), to)
	defer cancel()
	if err := s.svr.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) ListenAndServe() error {
	s.log.Infof("Starting server at %s", s.svr.Addr)
	var err error
	if s.svr.TLSConfig != nil {
		// this is already setup in the New, empties are ok here
		err = s.svr.ListenAndServeTLS("", "")
	} else {
		err = s.svr.ListenAndServe()
	}
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

type apiFunc struct {
	f func(router.Router)
}

func (a apiFunc) AddRoutes(r router.Router) {
	a.f(r)
}

func APIFunc(f func(router.Router)) APIDefinition {
	return apiFunc{f}
}
