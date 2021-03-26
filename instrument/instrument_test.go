package instrument

import (
	"reflect"
	"testing"

	"github.com/netlify/netlify-commons/testutil"
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

	require.NoError(t, mock.identify("myuser", analytics.NewTraits().SetName("My User")))
}

func TestLogging(t *testing.T) {
	cfg := Config{
		Key: "ABCD",
	}

	log, hook := testutil.TestLogger(t)

	client, err := NewClient(&cfg, log.WithField("component", "segment"))
	require.NoError(t, err)
	require.NoError(t, client.identify("myuser", analytics.NewTraits().SetName("My User")))
	assert.NotEmpty(t, hook.LastEntry())
}
