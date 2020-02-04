package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
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
	api APIDefinition
}

type Config struct {
	HealthPath string `split_words:"true"`
	Port       int
	TLS        nconf.TLSConfig
}

// APIDefinition is used to control lifecycle of the API
type APIDefinition interface {
	Start(r router.Router) error
	Stop()
}

func New(log logrus.FieldLogger, projectName string, config Config, api APIDefinition) (*Server, error) {
	r := router.New(
		log,
		router.OptHealthCheck(config.HealthPath, nil),
		router.OptTracingMiddleware(log, projectName),
	)

	if err := api.Start(r); err != nil {
		return nil, errors.Wrap(err, "Failed to start API")
	}

	s := Server{
		log: log.WithField("component", "server"),
		svr: &http.Server{
			Addr:    fmt.Sprintf(":%d", config.Port),
			Handler: r,
		},
		api: api,
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

	// Now that server is no longer listening, shutdown the API
	s.log.Info("Listener shutdown, stopping API")

	s.api.Stop()

	s.log.Debug("Completed shutting down the underlying API")

	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (s *Server) TestServer() *httptest.Server {
	return httptest.NewServer(s.svr.Handler)
}

type apiFunc struct {
	start func(router.Router) error
	stop  func()
}

func (a apiFunc) Start(r router.Router) error {
	return a.start(r)
}

func (a apiFunc) Stop() {
	a.stop()
}

func APIFunc(start func(router.Router) error, stop func()) APIDefinition {
	return apiFunc{
		start: start,
		stop:  stop,
	}
}
