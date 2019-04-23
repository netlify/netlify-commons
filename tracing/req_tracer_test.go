package tracing

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestTracerLogging(t *testing.T) {
	// func NewTracer(w http.ResponseWriter, r *http.Request, log logrus.FieldLogger, service string) (http.ResponseWriter, *http.Request, *RequestTracer) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://whatever.com/something", nil)

	log, hook := logtest.NewNullLogger()

	_, r, rt := NewTracer(rec, req, log, t.Name())

	rt.Start()
	e := hook.LastEntry()
	assert.Equal(t, 5, len(e.Data))
	assert.NotEmpty(t, e.Data["request_id"])
	assert.NotEmpty(t, e.Data["remote_addr"])
	assert.Empty(t, e.Data["referrer"])
	assert.NotEmpty(t, e.Data["method"])
	assert.Equal(t, "http://whatever.com/something", e.Data["url"])
	fmt.Println(e.Data)

	_ = SetLogField(r, "first", "second")
	SetFinalField(r, "final", "line").Info("should have the final here")
	e = hook.LastEntry()
	assert.Equal(t, 3, len(e.Data))
	assert.NotEmpty(t, e.Data["request_id"])
	assert.Equal(t, "line", e.Data["final"])
	assert.Equal(t, "second", e.Data["first"])
	fmt.Println(e.Data)

	rt.Info("Shouldn't have the final line")
	e = hook.LastEntry()
	assert.Equal(t, 2, len(e.Data))
	assert.NotEmpty(t, e.Data["request_id"])
	assert.Equal(t, "second", e.Data["first"])
	fmt.Println(e.Data)

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
