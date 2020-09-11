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
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
)

func TestHandleError_ErrorIsNil(t *testing.T) {
	logger, loggerOutput := test.NewNullLogger()
	w, r, _ := tracing.NewTracer(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/", nil),
		logger,
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
	)

	h := func(_ http.ResponseWriter, _ *http.Request) *HTTPError {
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
	assert.Equal(t, httpErr.InternalMessage, loggerOutput.AllEntries()[0].Message)
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
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	HandleError(errors.New("this is an error"), w, r)
	assert.Equal(t, 1, called)
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
func TestHandleError_ErrorFinishesTracer(t *testing.T) {
	logger, hook := test.NewNullLogger()
	w, r, _ := tracing.NewTracer(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/", nil),
		logger,
		t.Name(),
	)

	HandleError(errors.New("this is an error"), w, r)

	assert.Len(t, hook.Entries, 2)
	for _, e := range hook.Entries {
		switch e.Message {
		case "Completed Request":
		case "Unhandled server error: this is an error":
		default:
			assert.Fail(t, "unexpected message", e.Message)
		}
	}
}

func TestHandleError_ErrorFinishesSpan(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	mtracer := mocktracer.New()
	span := mtracer.StartSpan(t.Name())
	ctx := opentracing.ContextWithSpan(r.Context(), span)
	r = r.WithContext(ctx)
	r.Header.Set(tracing.HeaderRequestUUID, "123456")

	HandleError(errors.New("this is an error"), w, r)

	finished := mtracer.FinishedSpans()
	assert.Len(t, finished, 1)

	tags := finished[0].Tags()
	assert.Len(t, tags, 3)
	assert.Equal(t, "123456", tags["error_id"])
	assert.Equal(t, "this is an error", tags[ext.ErrorMsg])
	assert.Equal(t, 500, tags[ext.HTTPCode])
}
