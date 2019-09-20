package banlist

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestBanlistMissingFile(t *testing.T) {
	bl := newBanlist(tl(t), "not a path")
	require.Error(t, bl.update())
}

func TestBanlistInvalidFileContents(t *testing.T) {
	path, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(path.Name())

	_, err = path.WriteString("this isn't valid json")
	require.NoError(t, err)

	bl := newBanlist(tl(t), path.Name())
	require.Error(t, bl.update())
}

func TestBanlistNoPaths(t *testing.T) {
	bl := testList(t, &Config{
		Domains: []string{"something.com"},
	})

	assert.Len(t, bl.urls(), 0)
	domains := bl.domains()
	assert.Len(t, domains, 1)
	_, ok := domains["something.com"]
	assert.True(t, ok)
}

func TestBanlistNoDomains(t *testing.T) {
	bl := testList(t, &Config{
		URLs: []string{"something.com/path/to/thing"},
	})

	urls := bl.urls()
	assert.Len(t, urls, 1)
	_, ok := urls["something.com/path/to/thing"]
	assert.True(t, ok)

	assert.Len(t, bl.domains(), 0)
}
func TestBanlistBanning(t *testing.T) {
	bl := testList(t, &Config{
		URLs:    []string{"villians.com/the/joker"},
		Domains: []string{"sick.com"},
	})

	tests := []struct {
		url      string
		isBanned bool
		name     string
	}{
		{"http://heros.com", false, "completely unbanned"},
		{"http://sick.com:12345", true, "banned domain with port"},
		{"http://sick.com", true, "banned domain without port"},
		{"http://siCK.com", true, "banned domain mixed case"},
		{"http://villians.com:12354/the/joker", true, "banned path with port"},
		{"http://villians.com/the/joker", true, "banned path without port"},
		{"http://villians.com/the/Joker", true, "banned path mixed case"},
		{"http://villians.com/the/joker?query=param", true, "banned path with query params"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, test.url, nil)
			assert.Equal(t, test.isBanned, bl.CheckRequest(req))
		})
	}
}

func tl(t *testing.T) logrus.FieldLogger {
	return logrus.WithField("test", t.Name())
}

func testList(t *testing.T, config *Config) *Banlist {
	path, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	defer os.Remove(path.Name())

	require.NoError(t, json.NewEncoder(path).Encode(config))

	bl := newBanlist(tl(t), path.Name())
	require.NoError(t, bl.update())
	return bl
}
