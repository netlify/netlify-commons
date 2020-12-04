package tracing

import (
	"net/http"
	"strconv"

	opentracing "github.com/opentracing/opentracing-go"
	ext "github.com/opentracing/opentracing-go/ext"
	"github.com/sirupsen/logrus"
	ddtrace_ext "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentracer"
)

func TrackRequest(w http.ResponseWriter, r *http.Request, log logrus.FieldLogger, service, resource string, next http.Handler) {
	w, r, rt := NewTracer(w, r, log, service, resource)
	rt.Start()
	next.ServeHTTP(w, r)
	rt.Finish()
}

func WrapWithSpan(r *http.Request, reqID, service, resource string) (*http.Request, opentracing.Span) {
	span := opentracing.SpanFromContext(r.Context())
	if span != nil {
		return r, span
	}

	clientContext, _ := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span, ctx := opentracing.StartSpanFromContext(r.Context(), "http.handler",
		ext.RPCServerOption(clientContext),
		opentracer.ServiceName(service),
		opentracer.ResourceName(resource),
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
