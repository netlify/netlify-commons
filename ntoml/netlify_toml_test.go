package ntoml

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadingExample(t *testing.T) {
	tmp := testToml(t)
	defer os.Remove(t.Name())

	conf, err := LoadTomlFrom(tmp.Name())
	require.NoError(t, err)

	expected := &NetlifyToml{
		Settings: Settings{
			ID:   "this-is-a-site",
			Path: ".",
		},
		Redirects: []Redirect{
			{Origin: "/other", Destination: "/otherpage.html", Force: true},
		},
		Context: map[string]DeployContext{
			"deploy-preview": {
				Command: "hugo version && npm run build-preview",
			},
			"branch-deploy": {
				Command:     "hugo version && npm run build-branch",
				Environment: map[string]string{"HUGO_VERSION": "0.20.5"},
			},
		},
	}

	assert.Equal(t, expected, conf)
}

func TestSaveTomlFile(t *testing.T) {
	conf := &NetlifyToml{
		Settings: Settings{ID: "This is something", Path: "/dist"},
	}

	tmp, err := ioutil.TempFile("", "netlify-ctl")
	require.NoError(t, err)

	require.NoError(t, SaveTomlTo(conf, tmp.Name()))

	data, err := ioutil.ReadFile(tmp.Name())
	require.NoError(t, err)

	expected := `
	[settings]
	id: "This is something"
	path: "/dist"
	`

	assert.Equal(t, expected, string(data))
}

func testToml(t *testing.T) *os.File {
	tmp, err := ioutil.TempFile("", "netlify-ctl")
	require.NoError(t, err)

	data := `
[Settings]
  id = "this-is-a-site"
  path = "."

[[redirects]]
  origin = "/other"
	force = true
  destination = "/otherpage.html"

  [context.deploy-preview]
  command = "hugo version && npm run build-preview"


[context.branch-deploy]
  command = "hugo version && npm run build-branch"

  [context.branch-deploy.environment]
    HUGO_VERSION = "0.20.5"
`
	require.NoError(t, ioutil.WriteFile(tmp.Name(), []byte(data), 0664))

	return tmp
}
