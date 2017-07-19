package graceful

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartAndStop(t *testing.T) {
	orTimeout := func(c chan bool, i int, msg string) {
		select {
		case <-c:
		case <-time.After(time.Duration(i) * time.Second):
			assert.FailNow(t, msg)
		}
	}

	gotRequest := make(chan bool)
	clearRequest := make(chan bool)
	stoppedServer := make(chan bool)
	finishedListening := make(chan bool)

	var finished bool

	oh := func(w http.ResponseWriter, r *http.Request) {
		// trigger that we got the request
		close(gotRequest)
		// wait for a clear on that request
		orTimeout(clearRequest, 2, "waiting for request to be cleared")
		finished = true
	}

	svr := NewGracefulServer(http.HandlerFunc(oh), logrus.WithField("testing", true))
	require.NoError(t, svr.Bind("127.0.0.1:0"))

	go func() {
		assert.True(t, http.ErrServerClosed == svr.Listen())
		close(finishedListening)
	}()

	// make a request
	go func() {
		rsp, err := http.Get(svr.URL + "/something")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rsp.StatusCode)
	}()

	// wait for the origin to get the request
	orTimeout(gotRequest, 1, "didn't get the original request in time")

	// initate a shutdown
	go func() {
		svr.Close()
		close(stoppedServer)
	}()

	<-time.After(time.Second)

	// make a second request ~ should be bounced
	rsp, err := http.Get(svr.URL + "/something")
	switch e := err.(type) {
	case *url.Error:
		assert.True(t, strings.Contains(e.Error(), "connection refused"))
	default:
		assert.Fail(t, fmt.Sprintf("unknown type: %v", reflect.TypeOf(err)))
	}
	assert.Nil(t, rsp)

	// finish the first request
	close(clearRequest)

	// wait for server to close
	orTimeout(stoppedServer, 1, "didn't stop server in time")

	assert.True(t, finished)
	orTimeout(finishedListening, 1, "didn't actually stop the server in time")
}
