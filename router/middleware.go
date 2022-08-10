package router

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/bugsnag/bugsnag-go/v2"
	"github.com/netlify/netlify-commons/tracing"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
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
func Recoverer(errLog logrus.FieldLogger) Middleware {
	return MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		reqID := tracing.GetRequestID(r)

		defer func() {
			if rvr := recover(); rvr != nil {
				if errLog == nil {
					logger := logrus.New()
					logger.Out = os.Stderr
					errLog = logrus.NewEntry(logger)
				}
				panicLog := errLog.WithField("request_id", reqID)

				stack := debug.Stack()
				scanner := bufio.NewScanner(bytes.NewReader(stack))

				var lineID int
				panicLog.WithField("trace_line", lineID).Errorf("Panic: %+v", rvr)
				for scanner.Scan() {
					lineID++
					panicLog.WithField("trace_line", lineID).Errorf(scanner.Text())
				}

				se := &HTTPError{
					Code:    http.StatusInternalServerError,
					Message: http.StatusText(http.StatusInternalServerError),
				}
				HandleError(se, w, r)

				// in the event of a panic none of the normal shutdown code is called
				if span := opentracing.SpanFromContext(r.Context()); span != nil {
					span.SetTag("error_id", reqID)
					span.SetTag(ext.ErrorType, "panic")
					span.SetTag(ext.HTTPCode, http.StatusInternalServerError)

					if err, ok := rvr.(error); ok {
						span.SetTag(ext.ErrorMsg, err.Error())
					}

					defer span.Finish()
				}

				if tr := tracing.GetFromContext(r.Context()); tr != nil {
					tr.Finish()
				}
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

func BugSnag() Middleware {
	return MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		ctx := bugsnag.StartSession(r.Context())
		defer bugsnag.AutoNotify(ctx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func TrackAllRequests(log logrus.FieldLogger, service string) Middleware {
	return MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		// This is to maintain some legacy span work. It will cause the APM requests
		// to show up as the method on the top level
		tracing.TrackRequest(w, r, log, service, r.Method, true, next)
	})
}
