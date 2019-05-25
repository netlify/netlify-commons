package http

import (
	"testing"

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
