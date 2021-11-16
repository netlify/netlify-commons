package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/netlify/netlify-commons/nconf"
	"github.com/netlify/netlify-commons/router"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	defaultPort       = 9090
	defaultHealthPath = "/health"
)

// Server handles the setup and shutdown of the http server
// for an API
type Server struct {
	log      logrus.FieldLogger
	svr      *http.Server
	api      APIDefinition
	done     chan (bool)
	doneOnce sync.Once
}

type Config struct {
	HealthPath string `split_words:"true"`
	Port       int
	Host       string
	TLS        nconf.TLSConfig
}

// APIDefinition is used to control lifecycle of the API
type APIDefinition interface {
	Start(r router.Router) error
	Stop()
	Info() APIInfo
}

// APIInfo outlines the basic service information needed
type APIInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// HealthChecker is used to run a custom health check
// Implement it on your API if you want it to be checked
// when the healthcheck is called
type HealthChecker interface {
	Healthy(w http.ResponseWriter, r *http.Request) error
}

// NewOpts will create the server with many defaults. You can use the opts to override them.
// the one major default you can't change by this is the health path. This is set to /health
// and be enabled.
func NewOpts(log logrus.FieldLogger, api APIDefinition, opts ...Opt) (*Server, error) {
	defaultOpts := []Opt{
		WithHostAndPort("", defaultPort),
	}

	return buildServer(log, api, append(defaultOpts, opts...), defaultHealthPath)
}

// New will build a server with the defaults in place
func New(log logrus.FieldLogger, config Config, api APIDefinition) (*Server, error) {
	opts := []Opt{
		WithHostAndPort(config.Host, config.Port),
	}

	if config.TLS.Enabled {
		tcfg, err := config.TLS.TLSConfig()
		if err != nil {
			return nil, errors.Wrap(err, "Failed to build TLS config")
		}
		log.Info("TLS enabled")
		opts = append(opts, WithTLS(tcfg))
	}

	return buildServer(log, api, opts, config.HealthPath)
}

func (s *Server) Shutdown(to time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), to)
	defer cancel()
	defer func() {
		s.doneOnce.Do(func() {
			close(s.done)
		})
	}()

	if err := s.svr.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) ListenAndServe() error {
	go s.waitForShutdown()

	s.log.Infof("Starting server at %s", s.svr.Addr)
	var err error
	if s.svr.TLSConfig != nil {
		// this is already setup in the New, empties are ok here
		err = s.svr.ListenAndServeTLS("", "")
	} else {
		err = s.svr.ListenAndServe()
	}
	// Now that server is no longer listening
	s.log.Info("Listener shutdown, waiting for connections to drain")

	// Wait until Shutdown returns
	<-s.done

	s.log.Info("Connections are drained, shutting down API")

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

func (s *Server) waitForShutdown() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	s.log.Debug("Waiting for the shutdown signal")
	sig := <-sigs
	s.log.Infof("Received signal '%s', shutting down", sig)
	if err := s.Shutdown(30 * time.Second); err != nil {
		s.log.WithError(err).Warn("Failed to shutdown the server in time")
	}
}

type apiFunc struct {
	start func(router.Router) error
	stop  func()
	info  APIInfo
}

func (a apiFunc) Start(r router.Router) error {
	return a.start(r)
}

func (a apiFunc) Stop() {
	a.stop()
}
func (a apiFunc) Info() APIInfo {
	return a.info
}

func APIFunc(start func(router.Router) error, stop func(), info APIInfo) APIDefinition {
	return apiFunc{
		start: start,
		stop:  stop,
		info:  info,
	}
}

func buildRouter(log logrus.FieldLogger, api APIDefinition, healthPath string) router.Router {
	var healthHandler router.APIHandler
	if checker, ok := api.(HealthChecker); ok {
		healthHandler = checker.Healthy
	}

	r := router.New(
		log,
		router.OptHealthCheck(healthPath, healthHandler),
		router.OptEnableTracing(api.Info().Name),
		router.OptVersionHeader(api.Info().Name, api.Info().Version),
		router.OptRecoverer(),
	)

	return r
}

func buildServer(log logrus.FieldLogger, api APIDefinition, opts []Opt, healthPath string) (*Server, error) {
	r := buildRouter(log, api, healthPath)

	if err := api.Start(r); err != nil {
		return nil, errors.Wrap(err, "Failed to start API")
	}

	svr := new(http.Server)
	for _, o := range opts {
		o(svr)
	}
	svr.Handler = r
	s := Server{
		log:  log.WithField("component", "server"),
		svr:  svr,
		api:  api,
		done: make(chan bool),
	}
	return &s, nil
}
