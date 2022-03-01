package router

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/bugsnag/bugsnag-go/v2"
	"github.com/netlify/netlify-commons/metriks"
	"github.com/netlify/netlify-commons/tracing"
	"github.com/sirupsen/logrus"
)

// HTTPError is an error with a message and an HTTP status code.
type HTTPError struct {
	Code            int           `json:"code"`
	Message         string        `json:"msg"`
	JSON            interface{}   `json:"json,omitempty"`
	InternalError   error         `json:"-"`
	InternalMessage string        `json:"-"`
	ErrorID         string        `json:"error_id,omitempty"`
	Fields          logrus.Fields `json:"-"`
}

// BadRequestError creates a 400 HTTP error
func BadRequestError(fmtString string, args ...interface{}) *HTTPError {
	return httpError(http.StatusBadRequest, fmtString, args...)
}

// InternalServerError creates a 500 HTTP error
func InternalServerError(fmtString string, args ...interface{}) *HTTPError {
	return httpError(http.StatusInternalServerError, fmtString, args...)
}

// NotFoundError creates a 404 HTTP error
func NotFoundError(fmtString string, args ...interface{}) *HTTPError {
	return httpError(http.StatusNotFound, fmtString, args...)
}

// UnauthorizedError creates a 401 HTTP error
func UnauthorizedError(fmtString string, args ...interface{}) *HTTPError {
	return httpError(http.StatusUnauthorized, fmtString, args...)
}

// UnavailableServiceError creates a 503 HTTP error
func UnavailableServiceError(fmtString string, args ...interface{}) *HTTPError {
	return httpError(http.StatusServiceUnavailable, fmtString, args...)
}

// Error will describe the HTTP error in text
func (e *HTTPError) Error() string {
	if e.InternalMessage != "" {
		return e.InternalMessage
	}
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

// Cause will return the root cause error
func (e *HTTPError) Cause() error {
	if e.InternalError != nil {
		return e.InternalError
	}
	return e
}

// WithJSONError will add json details to the error
func (e *HTTPError) WithJSONError(json interface{}) *HTTPError {
	e.JSON = json
	return e
}

// WithInternalError will add internal error information to an error
func (e *HTTPError) WithInternalError(err error) *HTTPError {
	e.InternalError = err
	return e
}

// WithInternalMessage will add and internal message to an error
func (e *HTTPError) WithInternalMessage(fmtString string, args ...interface{}) *HTTPError {
	e.InternalMessage = fmt.Sprintf(fmtString, args...)
	return e
}

// WithFields will add fields to an error message
func (e *HTTPError) WithFields(fields logrus.Fields) *HTTPError {
	for key, value := range fields {
		e.Fields[key] = value
	}
	return e
}

// WithFields will add fields to an error message
func (e *HTTPError) WithField(key string, value interface{}) *HTTPError {
	e.Fields[key] = value
	return e
}

func httpError(code int, fmtString string, args ...interface{}) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: fmt.Sprintf(fmtString, args...),
		Fields:  make(logrus.Fields),
	}
}

// HandleError will handle any error. If it is of type *HTTPError then it will
// log anything of a 50x or an InternalError. It will write the right error response
// to the client. This way if you return a BadRequestError, it will simply write to the client.
// Any non-HTTPError will be treated as unhandled and result in a 50x
func HandleError(err error, w http.ResponseWriter, r *http.Request) {
	if err == nil || reflect.ValueOf(err).IsNil() {
		return
	}

	log := tracing.GetLogger(r)
	errorID := tracing.GetRequestID(r)

	var notifyBugsnag bool

	switch e := err.(type) {
	case *HTTPError:
		log = log.WithFields(e.Fields)

		e.ErrorID = errorID
		if e.Code >= http.StatusInternalServerError {
			notifyBugsnag = true
			elog := log.WithError(e)
			if e.InternalError != nil {
				elog = elog.WithField("internal_err", e.InternalError.Error())
			}

			elog.Errorf("internal server error: %s", e.InternalMessage)
		} else if e.InternalError != nil {
			notifyBugsnag = true
			log.WithError(e).Infof("unexpected error: %s", e.InternalMessage)
		}

		if jsonErr := SendJSON(w, e.Code, e); jsonErr != nil {
			log.WithError(jsonErr).Error("Failed to write the JSON error response")
		}
	default:
		notifyBugsnag = true
		metriks.Inc("unhandled_errors", 1)
		log.WithError(e).Errorf("Unhandled server error: %s", e.Error())
		// hide real error details from response to prevent info leaks
		w.WriteHeader(http.StatusInternalServerError)
		if _, writeErr := w.Write([]byte(`{"code":500,"msg":"Internal server error","error_id":"` + errorID + `"}`)); writeErr != nil {
			log.WithError(writeErr).Error("Error writing generic error message")
		}
	}

	if notifyBugsnag {
		bugsnag.Notify(err, r, r.Context(), bugsnag.MetaData{
			"meta": map[string]interface{}{
				"error_id":    errorID,
				"error_msg":   err.Error(),
				"status_code": http.StatusInternalServerError,
				"unhandled":   true,
			},
		})
	}
}
