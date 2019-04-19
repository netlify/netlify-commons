package tracing

import (
	"context"
	"net/http"
	"strconv"

	opentracing "github.com/opentracing/opentracing-go"
	ext "github.com/opentracing/opentracing-go/ext"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	ddtrace_ext "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentracer"
)

type contextKey string

const (
	HeaderRequestUUID = "X-BB-CLIENT-REQUEST-UUID"
)

func TrackRequest(w http.ResponseWriter, r *http.Request, log logrus.FieldLogger, service string, next http.Handler) {
	w, r, rt := NewTracer(w, r, log, service)
	rt.Start()
	next.ServeHTTP(w, r)
	rt.Finish()
}

func RequestID(r *http.Request) string {
	id := r.Header.Get(HeaderRequestUUID)
	if id == "" {
		id = uuid.NewV4().String()
		r.Header.Set(HeaderRequestUUID, id)
	}
	return id
}

func WrapWithLogger(r *http.Request, reqID string, log logrus.FieldLogger) (*http.Request, logrus.FieldLogger) {
	if r.Header.Get(HeaderNFDebugLogging) != "" {
		logger := logrus.New()
		logger.SetLevel(logrus.DebugLevel)

		if entry, ok := log.(*logrus.Entry); ok {
			log = logger.WithFields(entry.Data)
		}
	}

	log = log.WithFields(logrus.Fields{
		"method":      r.Method,
		"path":        r.URL.Path,
		"remote_addr": r.RemoteAddr,
		"referer":     r.Referer(),
		"request_id":  reqID,
	})

	entry := &structuredLoggerEntry{log}
	r = r.WithContext(context.WithValue(r.Context(), logKey, entry))

	return r, log
}

func WrapWithSpan(r *http.Request, reqID, service string) (*http.Request, opentracing.Span) {
	span := opentracing.SpanFromContext(r.Context())
	if span != nil {
		return r, span
	}

	clientContext, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span, ctx := opentracing.StartSpanFromContext(r.Context(), "http.handler",
		ext.RPCServerOption(clientContext),
		opentracer.ServiceName(service),
		opentracer.ResourceName(r.Method),
		opentracer.SpanType(ddtrace_ext.AppTypeWeb),
		opentracing.Tag{Key: "http.content_length", Value: strconv.FormatInt(r.ContentLength, 10)},
	)

	// datadog specific span.kind, normally "server"
	ext.Component.Set(span, "net/http")
	// "normal" is default request type until overridden
	ext.HTTPMethod.Set(span, r.Method)
	ext.HTTPUrl.Set(span, r.URL.String())
	scheme := "http"
	if r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	span.SetTag("http.base_url", scheme+"://"+r.Host)
	span.SetTag("http.request_id", reqID)
	return r.WithContext(ctx), span
}
