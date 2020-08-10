package featureflag

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	ld "gopkg.in/launchdarkly/go-server-sdk.v4"
)

// Ensure type checks
var _ Client = &ldClient{}
var _ Client = MockClient{}

func TestLDClient(t *testing.T) {
	config := ld.DefaultConfig
	config.Offline = true
	ldc, err := ld.MakeCustomClient("BAD_KEY", config, time.Second)
	require.NoError(t, err)

	client := ldClient{ldc, nil}
	isEnabled := client.Enabled("some_flag", "some_account_id")
	require.False(t, isEnabled)
}

func TestOfflineClient(t *testing.T) {
	client, err := NewClient("", 1*time.Second, &logrus.Entry{})
	require.NoError(t, err)

	require.False(t, client.Enabled("notset", "12345"))
	require.Equal(t, "foobar", client.Variation("notset", "foobar", "12345"))
}

func TestMockClient(t *testing.T) {
	mock := MockClient{
		BoolVars: map[string]bool{
			"FOO": true,
			"BAR": false,
		},
		StringVars: map[string]string{
			"FOO":  "BAR",
			"BLAH": "FOOBAR",
		},
	}

	require.True(t, mock.Enabled("FOO", "12345"))
	require.False(t, mock.Enabled("BAR", "12345"))
	require.False(t, mock.Enabled("NOTSET", "12345"))

	require.Equal(t, "BAR", mock.Variation("FOO", "DFLT", "12345"))
	require.Equal(t, "DFLT", mock.Variation("FOOBAR", "DFLT", "12345"))
}
