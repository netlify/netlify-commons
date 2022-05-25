package featureflag

import (
	"bytes"
	"testing"
	"time"

	"github.com/netlify/netlify-commons/util"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/launchdarkly/go-server-sdk.v5/ldfiledata"
)

func TestOfflineClient(t *testing.T) {
	cfg := Config{
		Key:            "ABCD",
		RequestTimeout: util.Duration{time.Second},
		Enabled:        false,
	}
	client, err := NewClient(&cfg, nil)
	require.NoError(t, err)

	require.False(t, client.Enabled("notset", "12345"))
	require.Equal(t, "foobar", client.Variation("notset", "foobar", "12345"))
	require.Equal(t, 3, client.Int("noset", 3, "12345"))
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
		IntVars: map[string]int{
			"FOO":  4,
			"BLAH": 7,
		},
	}

	require.True(t, mock.Enabled("FOO", "12345"))
	require.False(t, mock.Enabled("BAR", "12345"))
	require.False(t, mock.Enabled("NOTSET", "12345"))

	require.Equal(t, "BAR", mock.Variation("FOO", "DFLT", "12345"))
	require.Equal(t, "DFLT", mock.Variation("FOOBAR", "DFLT", "12345"))

	require.Equal(t, 4, mock.Int("FOO", 2, "12345"))
	require.Equal(t, 2, mock.Int("FOOBAR", 2, "12345"))
}

func TestAllEnabledFlags(t *testing.T) {
	fileSource := ldfiledata.DataSource().FilePaths("./fixtures/flags.yml")
	cfg := Config{
		Key:                    "ABCD",
		RequestTimeout:         util.Duration{time.Second},
		Enabled:                true,
		updateProcessorFactory: fileSource,
	}
	client, err := NewClient(&cfg, nil)
	require.NoError(t, err)

	flags := client.AllEnabledFlags("userid")

	require.Equal(t, []string{"my-boolean-flag-key"}, flags)
}

func TestLogging(t *testing.T) {
	cfg := Config{
		Key:            "ABCD",
		RequestTimeout: util.Duration{time.Second},
		Enabled:        false,
	}

	logBuf := new(bytes.Buffer)
	log := logrus.New()
	log.Out = logBuf

	_, err := NewClient(&cfg, log.WithField("component", "launch_darkly"))
	require.NoError(t, err)
	assert.NotEmpty(t, logBuf.Bytes())
}
