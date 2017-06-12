package transport

import (
	"errors"
	"testing"

	"github.com/netlify/netlify-commons/metrics"
	"github.com/stretchr/testify/assert"
)

func TestDoulbeSendNoError(t *testing.T) {
	var gotT1, gotT2 bool
	t1 := metrics.TransportFunc(func(m *metrics.RawMetric) error {
		gotT1 = true
		return nil
	}, nil)
	t2 := metrics.TransportFunc(func(m *metrics.RawMetric) error {
		gotT2 = true
		return nil
	}, nil)

	bt := NewBroadcastTransport([]metrics.Transport{t1, t2})

	assert.NoError(t, bt.Publish(&metrics.RawMetric{}))
	assert.True(t, gotT1)
	assert.True(t, gotT2)
}

func TestWriteAtTheSameTime(t *testing.T) {
	block := make(chan bool)
	t1 := metrics.TransportFunc(func(m *metrics.RawMetric) error {
		assert.Equal(t, int64(1), m.Value)
		// NOTE: we can modify in the goroutine...don't do it
		m.Value = 123
		block <- true
		return nil
	}, nil)
	t2 := metrics.TransportFunc(func(m *metrics.RawMetric) error {
		// this will block until the other one is called

		<-block
		assert.Equal(t, int64(123), m.Value)
		return nil
	}, nil)

	bt := NewBroadcastTransport([]metrics.Transport{t1, t2})
	m := &metrics.RawMetric{Value: 1}
	assert.NoError(t, bt.Publish(m))
	// can say this b/c order is enforced above...usually it is a race
	assert.Equal(t, int64(123), m.Value)
}

func TestModifyIsTogether(t *testing.T) {
	block := make(chan bool)
	var finallyCalled bool
	t1 := metrics.TransportFunc(func(m *metrics.RawMetric) error {
		block <- true
		return nil
	}, nil)
	t2 := metrics.TransportFunc(func(m *metrics.RawMetric) error {
		// this will block until the other one is called
		<-block
		finallyCalled = true
		return nil
	}, nil)

	bt := NewBroadcastTransport([]metrics.Transport{t1, t2})
	assert.NoError(t, bt.Publish(&metrics.RawMetric{Value: 1}))
	assert.True(t, finallyCalled)
}

func TestDoulbeSendWithError(t *testing.T) {
	t1 := metrics.TransportFunc(func(m *metrics.RawMetric) error {
		return errors.New("This is an error")
	}, nil)
	t2 := metrics.TransportFunc(func(m *metrics.RawMetric) error {
		return nil
	}, nil)

	bt := NewBroadcastTransport([]metrics.Transport{t1, t2})
	err := bt.Publish(&metrics.RawMetric{})
	if assert.NotNil(t, err) {
		ce, ok := err.(CompositeError)
		assert.True(t, ok)
		assert.Len(t, ce.errors, 1)
		assert.Equal(t, "This is an error", ce.Error())
	}
}
