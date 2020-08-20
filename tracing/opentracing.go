package tracing

import (
	"fmt"

	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func LogErrorToSpan(span opentracing.Span, err error) {
	if err == nil || span == nil {
		return
	}

	span.LogFields(otlog.String("event", "error"), otlog.Error(err))
	if sterr, ok := err.(stackTracer); ok {
		span.LogFields(otlog.String("stack", fmt.Sprintf("%+v", sterr.StackTrace())))
	}
}
