package tracing

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func TestGetFromContext_ContextContainsValue(t *testing.T) {
	tracer := new(RequestTracer)
	ctx := context.WithValue(context.Background(), tracerKey, tracer)

	assert.Same(t, tracer, GetFromContext(ctx))
}

func TestGetFromContext_ContextDoesNotContainValue(t *testing.T) {
	assert.Nil(t, GetFromContext(context.Background()))
}

func TestGetTracer_ContextContainsValue(t *testing.T) {
	tracer := new(RequestTracer)
	ctx := context.WithValue(context.Background(), tracerKey, tracer)

	r := httptest.NewRequest("", "/", nil)
	r = r.WithContext(ctx)

	assert.Same(t, tracer, GetTracer(r))
}

func TestGetTracer_ContextDoesNotContainValue(t *testing.T) {
	r := httptest.NewRequest("", "/", nil)

	assert.Nil(t, GetTracer(r))
}


