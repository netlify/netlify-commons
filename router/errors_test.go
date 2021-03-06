package router

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/armon/go-metrics"
	"github.com/bugsnag/bugsnag-go"
	"github.com/netlify/netlify-commons/metriks"
	"github.com/netlify/netlify-commons/tracing"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleError_ErrorIsNil(t *testing.T) {
	logger, loggerOutput := test.NewNullLogger()
	w, r, _ := tracing.NewTracer(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/", nil),
		logger,
		"test",
		"test",
	)

	HandleError(nil, w, r)

	assert.Empty(t, loggerOutput.AllEntries())
	assert.Empty(t, w.Header())
}

func TestHandleError_ErrorIsNilPointerToTypeHTTPError(t *testing.T) {
	logger, loggerOutput := test.NewNullLogger()
	w, r, _ := tracing.NewTracer(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/", nil),
		logger,
		"test",
		"test",
	)

	h := func(_ http.ResponseWriter, _ *http.Request) *HTTPError {
		return nil
	}

	HandleError(h(w, r), w, r)

	assert.Empty(t, loggerOutput.AllEntries())
	assert.Empty(t, w.Header())
}

func TestHandleError_ErrorIsNilInterface(t *testing.T) {
	logger, loggerOutput := test.NewNullLogger()
	w, r, _ := tracing.NewTracer(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/", nil),
		logger,
		"test",
		"test",
	)

	h := func(_ http.ResponseWriter, _ *http.Request) error {
		return nil
	}

	HandleError(h(w, r), w, r)

	assert.Empty(t, loggerOutput.AllEntries())
	assert.Empty(t, w.Header())
}

func TestHandleError_StandardError(t *testing.T) {
	logger, loggerOutput := test.NewNullLogger()
	w, r, _ := tracing.NewTracer(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/", nil),
		logger,
		"test",
		"test",
	)

	HandleError(errors.New("random error"), w, r)

	require.Len(t, loggerOutput.AllEntries(), 1)
	assert.Equal(t, "Unhandled server error: random error", loggerOutput.AllEntries()[0].Message)
	assert.Empty(t, w.Header())
}

func TestHandleError_HTTPError(t *testing.T) {
	logger, loggerOutput := test.NewNullLogger()
	recorder := httptest.NewRecorder()
	w, r, _ := tracing.NewTracer(
		recorder,
		httptest.NewRequest(http.MethodGet, "/", nil),
		logger,
		"test",
		"test",
	)

	httpErr := &HTTPError{
		Code:            http.StatusInternalServerError,
		Message:         http.StatusText(http.StatusInternalServerError),
		InternalError:   errors.New("random error"),
		InternalMessage: "Something unexpected happened",
	}

	HandleError(httpErr, w, r)

	resp := recorder.Result()
	b, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	expectedBody := fmt.Sprintf(`{"code":500,"msg":"Internal Server Error","error_id":"%s"}`, tracing.GetRequestID(r))
	assert.Equal(t, expectedBody, string(b))

	require.Len(t, loggerOutput.AllEntries(), 1)
	assert.Equal(t, "internal server error: "+httpErr.InternalMessage, loggerOutput.AllEntries()[0].Message)
}

func TestHandleError_NoLogForNormalErrors(t *testing.T) {
	logger, loggerOutput := test.NewNullLogger()
	recorder := httptest.NewRecorder()
	w, r, _ := tracing.NewTracer(
		recorder,
		httptest.NewRequest(http.MethodGet, "/", nil),
		logger,
		"test",
		"test",
	)

	httpErr := BadRequestError("not found yo.")

	HandleError(httpErr, w, r)

	resp := recorder.Result()
	b, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	expectedBody := fmt.Sprintf(`{"code":400,"msg":"not found yo.","error_id":"%s"}`, tracing.GetRequestID(r))
	assert.Equal(t, expectedBody, string(b))

	// we shouldn't log anything, this is a normal error
	require.Len(t, loggerOutput.AllEntries(), 0)
}

type OtherError struct {
	error string
}

func (e *OtherError) Error() string {
	return e.error
}

func TestHandleError_ErrorIsNilPointerToTypeOtherError(t *testing.T) {
	logger, loggerOutput := test.NewNullLogger()
	w, r, _ := tracing.NewTracer(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/", nil),
		logger,
		"test",
		"test",
	)

	var oe *OtherError

	HandleError(oe, w, r)

	require.Len(t, loggerOutput.AllEntries(), 0)
	assert.Empty(t, w.Header())
}

func TestHandleError_ErrorGoesToBugsnag(t *testing.T) {
	var called int

	bugsnag.OnBeforeNotify(func(event *bugsnag.Event, config *bugsnag.Configuration) error {
		called++
		require.NotNil(t, event)
		assert.NotNil(t, event.Ctx)

		assert.NotNil(t, config)
		return errors.New("this should stop us from sending to bugsnag")
	})

	HandleError(
		errors.New("this is an error"),
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/", nil),
	)
	assert.Equal(t, 1, called)

	// we shouldn't be notified of regular errors
	HandleError(
		NotFoundError("not found"),
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/", nil),
	)
	assert.Equal(t, 1, called)

	// we should be notified of internal server errors
	HandleError(
		InternalServerError("this is an error"),
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/", nil),
	)
	assert.Equal(t, 2, called)
}
func TestHandleError_ErrorEmitsMetric(t *testing.T) {
	sink := metrics.NewInmemSink(time.Minute, time.Minute)
	require.NoError(t, metriks.InitWithSink(t.Name(), sink))
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	HandleError(errors.New("this is an error"), w, r)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	assert.Len(t, sink.Data(), 1)
}
