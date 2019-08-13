package router

import (
	"github.com/netlify/netlify-commons/tracing"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

type Option func(r Router)

func OptEnableCORS(r Router) {
	corsMiddleware := cors.New(cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link", "X-Total-Count"},
		AllowCredentials: true,
	})
	r.Use(corsMiddleware.Handler)
}

func OptHealthCheck(path string, checker APIHandler) Option {
	return func(r Router) {
		r.Use(HealthCheck(path, checker))
	}
}

func OptVersionHeader(svcName, version string) Option {
	return func(r Router) {
		r.Use(VersionHeader(svcName, version))

	}
}

func OptTracingMiddleware(log logrus.FieldLogger, svcName string) Option {
	return func(r Router) {
		r.Use(tracing.Middleware(log, svcName))
	}
}


