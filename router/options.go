package router

import (
	"net/http"

	"github.com/netlify/netlify-commons/tracing"
)

type Option struct {
	f func(r *chiWrapper)
}

func OptEnableCORS() Option {
	return Option{
		f: func(r *chiWrapper) {
			r.enableCORS = true
		},
	}
}

func OptHealthCheck(path string, checker APIHandler) Option {
	return Option{
		f: func(r *chiWrapper) {
			r.healthEndpoint = path
			r.healthHandler = checker
		},
	}
}

func OptVersionHeader(svcName, version string) Option {
	return Option{
		f: func(r *chiWrapper) {
			if version == "" {
				version = "unknown"
			}
			r.version = version
			r.svcName = svcName
		},
	}
}

func OptEnableTracing(svcName string) Option {
	return Option{
		f: func(r *chiWrapper) {
			r.svcName = svcName
			r.enableTracing = true
			r.chi.NotFound(func(w http.ResponseWriter, req *http.Request) {
				tracing.TrackRequest(
					w,
					req,
					r.rootLogger,
					svcName,
					req.Method,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusNotFound)
					}),
				)
			})
			r.chi.MethodNotAllowed(func(w http.ResponseWriter, req *http.Request) {
				tracing.TrackRequest(
					w,
					req,
					r.rootLogger,
					svcName,
					req.Method,
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusNotFound)
					}),
				)
			})
		},
	}
}

func OptRecoverer() Option {
	return Option{
		f: func(r *chiWrapper) {
			r.enableRecover = true
		},
	}
}
