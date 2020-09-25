package featureflag

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOfflineClient(t *testing.T) {
	cfg := Config{
		Key:            "ABCD",
		RequestTimeout: time.Second,
		Enabled:        false,
	}
	client, err := NewClient(&cfg, nil)
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

func TestGetEnabledFlagsFromState(t *testing.T) {

}
