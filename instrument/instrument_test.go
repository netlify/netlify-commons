package instrument

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/netlify/netlify-commons/testutil"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/segmentio/analytics-go.v3"
)

func TestLogOnlyClient(t *testing.T) {
	cfg := Config{
		Key:     "ABCD",
		Enabled: false,
	}
	client, err := NewClient(&cfg, nil)
	require.NoError(t, err)

	require.Equal(t, reflect.TypeOf(&MockClient{}).Kind(), reflect.TypeOf(client).Kind())
}

func TestMockClient(t *testing.T) {
	log := testutil.TL(t)
	mock := MockClient{log}

	require.NoError(t, mock.Identify("myuser", analytics.NewTraits().SetName("My User")))
}

func TestLogging(t *testing.T) {
	cfg := Config{
		Key: "ABCD",
	}

	logBuf := new(bytes.Buffer)
	log := logrus.New()
	log.Out = logBuf

	client, err := NewClient(&cfg, log.WithField("component", "segment"))
	require.NoError(t, err)
	require.NoError(t, client.Identify("myuser", analytics.NewTraits().SetName("My User")))
	assert.NotEmpty(t, logBuf.Bytes())
}
