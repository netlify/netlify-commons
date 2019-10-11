package tracing

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func TestGetLoggerFromContext_ContextContainsValue(t *testing.T) {
	tracer := new(RequestTracer)
	tracer.FieldLogger = logrus.New()
	ctx := context.WithValue(context.Background(), tracerKey, tracer)

	assert.Same(t, tracer.FieldLogger, GetLoggerFromContext(ctx))
}

func TestGetLoggerFromContext_ContextDoesNotContainValue(t *testing.T) {
	l := GetLoggerFromContext(context.Background())
	assert.NotNil(t, l)
	assert.Implements(t, (*logrus.FieldLogger)(nil), l)
}

func TestGetLogger_ContextContainsValue(t *testing.T) {
	tracer := new(RequestTracer)
	tracer.FieldLogger = logrus.New()
	ctx := context.WithValue(context.Background(), tracerKey, tracer)

	r := httptest.NewRequest("", "/", nil)
	r = r.WithContext(ctx)

	assert.Same(t, tracer.FieldLogger, GetLogger(r))
}

func TestGetLogger_ContextDoesNotContainValue(t *testing.T) {
	r := httptest.NewRequest("", "/", nil)
	l := GetLogger(r)

	assert.NotNil(t, l)
	assert.Implements(t, (*logrus.FieldLogger)(nil), l)
}

