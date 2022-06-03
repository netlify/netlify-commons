package featureflag

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGlobalAccess(t *testing.T) {
	// initial value should be default
	require.Equal(t, defaultClient, GetGlobalClient())

	// setting new global should be reflected
	n := &ldClient{}
	SetGlobalClient(n)
	require.Equal(t, n, GetGlobalClient())
}
