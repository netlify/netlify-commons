package transport

import (
	"strings"
	"sync"

	"github.com/netlify/netlify-commons/metrics"
)

type BroadcastTransport struct {
	transports []metrics.Transport
}

type CompositeError struct {
	errors []error
}

func (e CompositeError) Error() string {
	if len(e.errors) == 0 {
		return "unknown error"
	}

	errMsgs := []string{}
	for _, err := range e.errors {
		errMsgs = append(errMsgs, err.Error())
	}

	return strings.Join(errMsgs, "\n")
}

func NewBroadcastTransport(transports []metrics.Transport) *BroadcastTransport {
	return &BroadcastTransport{transports}
}

func (t BroadcastTransport) Publish(m *metrics.RawMetric) error {
	errors := make(chan error, len(t.transports))
	wg := sync.WaitGroup{}

	for _, trans := range t.transports {
		wg.Add(1)
		go func(trans metrics.Transport) {
			errors <- trans.Publish(m)
			wg.Done()
		}(trans)
	}

	wg.Wait()
	close(errors)

	childErrors := []error{}
	for e := range errors {
		if e != nil {
			childErrors = append(childErrors, e)
		}
	}

	if len(childErrors) > 0 {
		return CompositeError{childErrors}
	}

	return nil
}
