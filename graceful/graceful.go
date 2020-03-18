package graceful

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// Shutdownable is a target that can be closed gracefully
type Shutdownable interface {
	Shutdown(context.Context) error
}

type target struct {
	name    string
	shut    Shutdownable
	timeout time.Duration
}

// Closer handles shutdown of servers and connections
type Closer struct {
	targets      []target
	targetsMutex sync.Mutex

	done     chan struct{}
	doneBool int32
}

// Register inserts a target to shutdown gracefully
func (cc *Closer) Register(name string, shut Shutdownable, timeout time.Duration) {
	cc.targetsMutex.Lock()
	cc.targets = append(cc.targets, target{
		name:    name,
		shut:    shut,
		timeout: timeout,
	})
	cc.targetsMutex.Unlock()
}

// DetectShutdown asynchronously waits for a shutdown signal and then shuts down gracefully
// Returns a function to trigger a shutdown from the outside, like cancelling a context
func (cc *Closer) DetectShutdown(log logrus.FieldLogger) func() {
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		select {
		case sig := <-signals:
			log.Infof("Triggering shutdown from signal %s", sig)
		case <-cc.done:
			log.Infof("Shutting down...")
		}

		if atomic.SwapInt32(&cc.doneBool, 1) != 1 {
			wg := sync.WaitGroup{}
			cc.targetsMutex.Lock()
			for _, targ := range cc.targets {
				wg.Add(1)
				go func(targ target, log logrus.FieldLogger) {
					defer wg.Done()

					ctx, cancel := context.WithTimeout(context.Background(), targ.timeout)
					defer cancel()

					if err := targ.shut.Shutdown(ctx); err != nil {
						log.WithError(err).Error("Graceful shutdown failed")
					} else {
						log.Info("Shutdown finished")
					}
				}(targ, log.WithField("target", targ.name))
			}
			cc.targetsMutex.Unlock()
			wg.Wait()
			os.Exit(0)
		}
	}()

	return func() {
		cc.done <- struct{}{}
	}
}
