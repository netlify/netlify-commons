package tracing

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

func Middleware(log logrus.FieldLogger, svcName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			TrackRequest(w, r, log, svcName, next)
		})
	}
}
