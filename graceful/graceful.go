package graceful

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

var DefaultShutdownTimeout = time.Second * 60

const shutdown uint32 = 1

type GracefulServer struct {
	server   *http.Server
	listener net.Listener
	log      *logrus.Entry

	exit chan struct{}

	URL             string
	state           uint32
	ShutdownTimeout time.Duration
	ShutdownError   error
}

func NewGracefulServer(handler http.Handler, log *logrus.Entry) *GracefulServer {
	return &GracefulServer{
		server:          &http.Server{Handler: handler},
		log:             log,
		listener:        nil,
		exit:            make(chan struct{}),
		ShutdownTimeout: DefaultShutdownTimeout,
	}
}

func (svr *GracefulServer) Bind(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	svr.URL = "http://" + l.Addr().String()
	svr.listener = l
	return nil
}

func (svr *GracefulServer) Listen() error {
	go svr.listenForShutdownSignal()
	serveErr := svr.server.Serve(svr.listener)
	if serveErr != http.ErrServerClosed {
		svr.log.WithError(serveErr).Warn("Error while running server")
		return serveErr
	}

	<-svr.exit

	return svr.ShutdownError
}

func (svr *GracefulServer) listenForShutdownSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	sig := <-c
	svr.log.Infof("Triggering shutdown from signal %s", sig)
	svr.Shutdown()
}

func (svr *GracefulServer) ListenAndServe(addr string) error {
	if svr.listener != nil {
		return errors.New("The listener has already started, don't call Bind first")
	}
	if err := svr.Bind(addr); err != nil {
		return err
	}

	return svr.Listen()
}

func (svr *GracefulServer) Shutdown() error {
	if atomic.SwapUint32(&svr.state, shutdown) == shutdown {
		svr.log.Debug("Calling shutdown on already shutdown server, ignoring")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), svr.ShutdownTimeout)
	defer cancel()

	svr.log.Infof("Triggering shutdown, in at most %s ", svr.ShutdownTimeout.String())
	shutErr := svr.server.Shutdown(ctx)
	if shutErr == context.DeadlineExceeded {
		svr.log.WithError(shutErr).Warnf("Forcing a shutdown after waiting %s", svr.ShutdownTimeout.String())
		shutErr = svr.server.Close()
	}

	svr.ShutdownError = shutErr
	close(svr.exit)

	return shutErr
}
