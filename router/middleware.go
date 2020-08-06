package router

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/netlify/netlify-commons/tracing"
	"github.com/sirupsen/logrus"
)

var bearerRegexp = regexp.MustCompile(`^(?:B|b)earer (\S+$)`)

const versionHeaderTempl = "X-NF-%s-Version"

type Middleware func(http.Handler) http.Handler

func MiddlewareFunc(f func(w http.ResponseWriter, r *http.Request, next http.Handler)) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			f(w, r, next)
		})
	}
}

func VersionHeader(serviceName, version string) Middleware {
	if version == "" {
		version = "unknown"
	}
	return MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		w.Header().Set(fmt.Sprintf(versionHeaderTempl, strings.ToUpper(serviceName)), version)
		next.ServeHTTP(w, r)
	})
}

func CheckAuth(secret string) Middleware {
	return MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		if secret != "" {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				HandleError(UnauthorizedError("This endpoint requires a Bearer token"), w, r)
				return
			}

			matches := bearerRegexp.FindStringSubmatch(authHeader)
			if len(matches) != 2 {
				HandleError(UnauthorizedError("This endpoint requires a Bearer token"), w, r)
				return
			}

			if secret != matches[1] {
				HandleError(UnauthorizedError("This endpoint requires a Bearer token"), w, r)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// Recoverer is a middleware that recovers from panics, logs the panic (and a
// backtrace), and returns a HTTP 500 (Internal Server Error) status if
// possible. Recoverer prints a request ID if one is provided.
func Recoverer(rootLogger logrus.FieldLogger) Middleware {
	return MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		defer func() {
			if rvr := recover(); rvr != nil {
				if log := tracing.GetLogger(r); log != nil {
					log.Errorf("Panic: %+v\n%s", rvr, debug.Stack())
				} else if rootLogger != nil {
					rootLogger.Errorf("Panic: %+v\n%s", rvr, debug.Stack())
				} else {
					fmt.Fprintf(os.Stderr, "Panic: %+v\n", rvr)
					debug.PrintStack()
				}

				se := &HTTPError{
					Code:    http.StatusInternalServerError,
					Message: http.StatusText(http.StatusInternalServerError),
				}
				HandleError(se, w, r)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func HealthCheck(route string, f APIHandler) Middleware {
	return MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		if r.URL.Path == route {
			if f == nil {
				w.WriteHeader(http.StatusOK)
				return
			}

			if err := f(w, r); err != nil {
				HandleError(err, w, r)
			}

			return
		}
		next.ServeHTTP(w, r)
	})
}
