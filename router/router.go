package router

import (
	"net/http"
	"strings"

	"github.com/netlify/netlify-commons/tracing"
	"github.com/rs/cors"

	"github.com/go-chi/chi"
	"github.com/sebest/xff"
	"github.com/sirupsen/logrus"
)

type chiWrapper struct {
	chi chi.Router

	version       string
	svcName       string
	tracingPrefix string
	rootLogger    logrus.FieldLogger

	healthEndpoint string
	healthHandler  APIHandler

	enableTracing bool
	enableCORS    bool
	enableRecover bool
}

// Router wraps the chi router to make it slightly more accessible
type Router interface {
	// Use appends one middleware onto the Router stack.
	Use(fn Middleware)

	// With adds an inline middleware for an endpoint handler.
	With(fn Middleware) Router

	// Route mounts a sub-Router along a `pattern`` string.
	Route(pattern string, fn func(r Router))

	// Method adds a routes for a `pattern` that matches the `method` HTTP method.
	Method(method, pattern string, h APIHandler)

	// HTTP-method routing along `pattern`
	Delete(pattern string, h APIHandler)
	Get(pattern string, h APIHandler)
	Post(pattern string, h APIHandler)
	Put(pattern string, h APIHandler)

	// Mount attaches another http.Handler along ./pattern/*
	Mount(pattern string, h http.Handler)

	ServeHTTP(http.ResponseWriter, *http.Request)
}

// New creates a router with sensible defaults (xff, request id, cors)
func New(log logrus.FieldLogger, options ...Option) Router {
	r := &chiWrapper{
		chi:        chi.NewRouter(),
		version:    "unknown",
		rootLogger: log,
	}

	xffmw, _ := xff.Default()
	r.Use(xffmw.Handler)
	for _, opt := range options {
		opt.f(r)
	}

	if r.enableRecover {
		r.Use(Recoverer(log))
	}
	r.Use(VersionHeader(r.svcName, r.version))

	// we don't want to track health requests, they're noise
	if r.healthEndpoint != "" {
		r.Use(HealthCheck(r.healthEndpoint, r.healthHandler))
	}

	// this needs to be in this order so that we can make sure
	// that the tracing middleware is at the root of the stack
	// other than version and recovery
	if r.enableTracing {
		r.Use(r.tracingMiddleware())
	}
	if r.enableCORS {
		corsMiddleware := cors.New(cors.Options{
			AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
			ExposedHeaders:   []string{"Link", "X-Total-Count"},
			AllowCredentials: true,
		})
		r.Use(corsMiddleware.Handler)
	}

	return r
}

// Route allows creating a generic route
func (r *chiWrapper) Route(pattern string, fn func(Router)) {
	r.chi.Route(pattern, func(c chi.Router) {
		wrapper := new(chiWrapper)
		*wrapper = *r
		wrapper.chi = c
		wrapper.tracingPrefix = sanitizePattern(pattern)
		fn(wrapper)
	})
}

// Method adds a routes for a `pattern` that matches the `method` HTTP method.
func (r *chiWrapper) Method(method, pattern string, h APIHandler) {
	r.chi.Method(method, pattern, r.traceRequest(method, pattern, h))
}

// Get adds a GET route
func (r *chiWrapper) Get(pattern string, fn APIHandler) {

	wrapper := new(chiWrapper)
	*wrapper = *r
	wrapper.tracingPrefix = sanitizePattern(pattern)

	wrapper.chi.Get(pattern, wrapper.traceRequest(http.MethodGet, pattern, fn))
}

// Post adds a POST route
func (r *chiWrapper) Post(pattern string, fn APIHandler) {
	r.chi.Post(pattern, r.traceRequest(http.MethodPost, pattern, fn))
}

// Put adds a PUT route
func (r *chiWrapper) Put(pattern string, fn APIHandler) {
	r.chi.Put(pattern, r.traceRequest(http.MethodPut, pattern, fn))
}

// Delete adds a DELETE route
func (r *chiWrapper) Delete(pattern string, fn APIHandler) {
	r.chi.Delete(pattern, r.traceRequest(http.MethodDelete, pattern, fn))
}

// WithBypass adds an inline chi middleware for an endpoint handler
func (r *chiWrapper) With(fn Middleware) Router {
	r.chi = r.chi.With(fn)
	return r
}

// UseBypass appends one chi middleware onto the Router stack
func (r *chiWrapper) Use(fn Middleware) {
	r.chi.Use(fn)
}

// ServeHTTP will serve a request
func (r *chiWrapper) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.chi.ServeHTTP(w, req)
}

// Mount attaches another http.Handler along ./pattern/*
func (r *chiWrapper) Mount(pattern string, h http.Handler) {
	if r.enableTracing {
		h = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			tracing.TrackRequest(w, req, r.rootLogger, r.svcName, pattern, h)
		})
	}
	r.chi.Mount(pattern, h)
}

// =======================================
// HTTP handler with custom error payload
// =======================================

type APIHandler func(w http.ResponseWriter, r *http.Request) error

func HandlerFunc(fn APIHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			HandleError(err, w, r)
		}
	}
}

func (r *chiWrapper) traceRequest(method, pattern string, fn APIHandler) http.HandlerFunc {
	f := HandlerFunc(fn)
	// if r.tracingMiddleware != nil {
	// 	r.tracingMiddleware.registerHandler(sanitizePattern(pattern), f)
	// pattern =
	// if r.tracingPrefix != "" {
	// 	pattern = r.tracingPrefix + "." + pattern
	// }
	// resourceName := strings.ToUpper(method)
	// if pattern != "" {
	// 	resourceName += "::" + pattern
	// }

	// child := &chiWrapper{}
	// *child = *r

	// return func(w http.ResponseWriter, req *http.Request) {
	// 	tracing.TrackRequest(w, req, r.rootLogger, r.svcName, resourceName, f)
	// }
	// }
	return f
}

func sanitizePattern(pattern string) string {
	pattern = strings.TrimPrefix(pattern, "/")
	pattern = strings.ReplaceAll(pattern, "{", "")
	pattern = strings.ReplaceAll(pattern, "}", "")
	pattern = strings.ReplaceAll(pattern, "/", ".")
	pattern = strings.TrimSuffix(pattern, ".")
	return pattern
}

func (r *chiWrapper) tracingMiddleware() Middleware {
	return MiddlewareFunc(func(w http.ResponseWriter, req *http.Request, next http.Handler) {
		resourceName := strings.ToUpper(req.Method)
		// if pattern != "" {
		// 	resourceName += "::" + pattern
		// }
		tracing.TrackRequest(w, req, r.rootLogger, r.svcName, resourceName, next)
	})
}
