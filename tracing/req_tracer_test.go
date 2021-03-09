package tracing

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
)

func TestTracerLogging(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://whatever.com/something", nil)

	log, hook := logtest.NewNullLogger()

	_, r, rt := NewTracer(rec, req, log, t.Name(), "some_resource")

	rt.Start()
	e := hook.LastEntry()
	assert.Equal(t, 5, len(e.Data))
	assert.NotEmpty(t, e.Data["request_id"])
	assert.NotEmpty(t, e.Data["remote_addr"])
	assert.Empty(t, e.Data["referrer"])
	assert.NotEmpty(t, e.Data["method"])
	assert.Equal(t, "http://whatever.com/something", e.Data["url"])

	_ = SetLogField(r, "first", "second")
	SetFinalField(r, "final", "line").Info("should have the final here")
	e = hook.LastEntry()
	assert.Equal(t, 3, len(e.Data))
	assert.NotEmpty(t, e.Data["request_id"])
	assert.Equal(t, "line", e.Data["final"])
	assert.Equal(t, "second", e.Data["first"])

	rt.Info("Shouldn't have the final line")
	e = hook.LastEntry()
	assert.Equal(t, 2, len(e.Data))
	assert.NotEmpty(t, e.Data["request_id"])
	assert.Equal(t, "second", e.Data["first"])

	rt.WriteHeader(http.StatusOK)
	rt.Write([]byte{0, 1, 2, 3})
	rt.Finish()
	e = hook.LastEntry()

	assert.Equal(t, 9, len(e.Data))

	// the automatic fields
	assert.NotEmpty(t, e.Data["dur"])
	assert.NotEmpty(t, e.Data["dur_ns"])
	assert.NotEmpty(t, e.Data["request_id"])
	assert.Equal(t, 4, e.Data["rsp_bytes"])
	assert.Equal(t, 200, e.Data["status_code"])
	assert.Equal(t, "http://whatever.com/something", e.Data["url"])
	assert.Equal(t, "GET", e.Data["method"])

	// the value that we added above
	assert.Equal(t, "second", e.Data["first"])
	assert.Equal(t, "line", e.Data["final"])
}

func TestTracerSpans(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://whatever.com/something", nil)

	log, _ := logtest.NewNullLogger()

	mt := mocktracer.New()
	opentracing.SetGlobalTracer(mt)
	_, _, rt := NewTracer(rec, req, log, t.Name(), "some_resource")
	rt.Start()
	rt.WriteHeader(http.StatusOK)
	rt.Finish()

	require.Len(t, mt.FinishedSpans(), 1)
	span := mt.FinishedSpans()[0]
	assert.Equal(t, t.Name(), span.Tag(ext.ServiceName))
	assert.Equal(t, "some_resource", span.Tag(ext.ResourceName))
	assert.Equal(t, http.MethodGet, span.Tag(ext.HTTPMethod))
	assert.Equal(t, strconv.Itoa(http.StatusOK), span.Tag(ext.HTTPCode))
	assert.Equal(t, rt.RequestID, span.Tag("http.request_id"))
}
