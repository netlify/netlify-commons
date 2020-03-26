package router

import (
	"fmt"
	"net/http"

	"github.com/netlify/netlify-commons/tracing"
)

// HTTPError is an error with a message and an HTTP status code.
type HTTPError struct {
	Code            int         `json:"code"`
	Message         string      `json:"msg"`
	JSON            interface{} `json:"json"`
	InternalError   error       `json:"-"`
	InternalMessage string      `json:"-"`
	ErrorID         string      `json:"error_id,omitempty"`
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

func httpError(code int, fmtString string, args ...interface{}) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: fmt.Sprintf(fmtString, args...),
	}
}

// HandleError will handle an error
func HandleError(err error, w http.ResponseWriter, r *http.Request) {
	log := tracing.GetLogger(r)
	errorID := tracing.RequestID(r)
	switch e := err.(type) {
	case nil:
		return
	case *HTTPError:
		// assert to *HTTPError to check nil intrface
		httpErr := err.(*HTTPError)
		if httpErr == nil {
			return
		}
		if e.Code >= http.StatusInternalServerError {
			e.ErrorID = errorID
			// this will get us the stack trace too
			log.WithError(e.Cause()).Error(e.Error())
		} else {
			log.WithError(e.Cause()).Info(e.Error())
		}

		if jsonErr := SendJSON(w, e.Code, e); jsonErr != nil {
			HandleError(jsonErr, w, r)
		}
	default:
		// do not call the Error() method, this will cause a panic if a custom error is passed in is a nil interface
		// instead rely on the fmt package to look for the error interface when printing values
		log.WithError(e).Errorf("Unhandled server error: %s", e)
		// hide real error details from response to prevent info leaks
		w.WriteHeader(http.StatusInternalServerError)
		if _, writeErr := w.Write([]byte(`{"code":500,"msg":"Internal server error","error_id":"` + errorID + `"}`)); writeErr != nil {
			log.WithError(writeErr).Error("Error writing generic error message")
		}
	}
}
