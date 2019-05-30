package http

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestIsLocalAddress(t *testing.T) {
	assert.False(t, isLocalAddress("216.58.194.206"))
	assert.True(t, isLocalAddress("127.0.0.1"))
	assert.True(t, isLocalAddress("10.0.0.1"))
	assert.True(t, isLocalAddress("192.168.0.1"))
	assert.True(t, isLocalAddress("172.16.0.0"))
	assert.True(t, isLocalAddress("169.254.169.254"))
}

type countServer struct {
	count int
}

func (c *countServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.count++
	w.Write([]byte("Done"))
}

func TestSafeHTTPClient(t *testing.T) {
	client := &http.Client{}

	counter := &countServer{}
	s := &http.Server{
		Addr:    ":9999",
		Handler: counter,
	}
	go func() {
		s.ListenAndServe()
	}()

	res, err := client.Get("http://localhost:9999")
	assert.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, 1, counter.count)

	client = SafeHTTPClient(client, logrus.New())

	res, err = client.Get("http://localhost:9999")
	assert.NotNil(t, err)
	assert.Nil(t, res)
	assert.Equal(t, 1, counter.count)

	fmt.Println("Closing down")
	s.Shutdown(context.TODO())
}
