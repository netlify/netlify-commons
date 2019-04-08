package tracing

import (
	"bufio"
	"fmt"
	"net"
	"net/http"

	"github.com/sirupsen/logrus"
)

type trackingWriter struct {
	writer   http.ResponseWriter
	rspBytes int
	written  bool
	status   int
	log      logrus.FieldLogger
}

func (w *trackingWriter) Write(in []byte) (int, error) {
	w.rspBytes += len(in)
	return w.writer.Write(in)
}

func (w *trackingWriter) WriteHeader(code int) {
	if w.written {
		w.log.Warnf("Attempted to write the header twice: %d first, %d second", w.status, code)
		return
	}
	w.status = code
	w.written = true
	w.writer.WriteHeader(code)
}

func (w *trackingWriter) Header() http.Header {
	return w.writer.Header()
}

func (w *trackingWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := w.writer.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("webserver doesn't support hijacking")
	}
	return hj.Hijack()
}
