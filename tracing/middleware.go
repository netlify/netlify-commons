package tracing

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

func NewTracingMiddleware(log logrus.FieldLogger, service string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		TrackRequest(w, r, log, service, next)
	})
}
