package router

import (
	"errors"
	"fmt"
	"github.com/netlify/netlify-commons/tracing"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
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

func TestHandleError_StandardError(t *testing.T) {
	logger, loggerOutput := test.NewNullLogger()
	w, r, _ := tracing.NewTracer(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/", nil),
		logger,
		"test",
	)

	HandleError(errors.New("random error"), w, r)

	require.Len(t,  loggerOutput.AllEntries(), 1)
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
		Code: http.StatusInternalServerError,
		Message: http.StatusText(http.StatusInternalServerError),
		InternalError: errors.New("random error"),
		InternalMessage: "Something unexpected happened",
	}

	HandleError(httpErr, w, r)

	resp := recorder.Result()
	b, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	expectedBody := fmt.Sprintf(`{"code":500,"msg":"Internal Server Error","json":null,"error_id":"%s"}`, tracing.RequestID(r))
	assert.Equal(t, expectedBody, string(b))

	require.Len(t,  loggerOutput.AllEntries(), 1)
	assert.Equal(t, httpErr.InternalMessage, loggerOutput.AllEntries()[0].Message)
}

