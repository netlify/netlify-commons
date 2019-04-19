package tracing

import (
	"net/http"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

type RequestTracer struct {
	*trackingWriter
	RequestID string

	request *http.Request
	span    opentracing.Span
	start   time.Time
}

func NewTracer(w http.ResponseWriter, r *http.Request, log logrus.FieldLogger, service string) (http.ResponseWriter, *http.Request, *RequestTracer) {
	reqID := RequestID(r)

	r, log = WrapWithLogger(r, reqID, log)
	r, span := WrapWithSpan(r, reqID, service)
	trackWriter := &trackingWriter{
		writer: w,
		log:    log,
	}

	rt := &RequestTracer{
		RequestID:      reqID,
		span:           span,
		trackingWriter: trackWriter,
		request:        r,
	}

	return rt, r, rt
}

func (rt *RequestTracer) Start() {
	rt.start = time.Now()
	rt.Log().WithField("url", rt.request.URL.String()).Info("Starting Request")
}

func (rt *RequestTracer) Finish() {
	dur := time.Since(rt.start)
	fields := logrus.Fields{
		"status_code": rt.trackingWriter.status,
		"rsp_bytes":   rt.trackingWriter.rspBytes,
		"dur":         dur.String(),
		"dur_ns":      dur.Nanoseconds(),
	}
	rt.span.Finish()
	rt.Log().WithFields(fields).Info("Completed Request")
}

func (rt *RequestTracer) Log() logrus.FieldLogger {
	return GetLogger(rt.request)
}

func (rt *RequestTracer) SetLogField(key string, value interface{}) logrus.FieldLogger {
	return SetLogField(rt.request, key, value)
}

func (rt *RequestTracer) SetLogFields(fields logrus.Fields) logrus.FieldLogger {
	return SetLogFields(rt.request, fields)
}
