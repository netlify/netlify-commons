package nconf

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	Hero     string
	Villian  string
	Matchups map[string]string
	Cities   []string

	ShootingLocation string `mapstructure:"shooting_location" split_words:"true"`
}

func exampleConfig() testConfig {
	return testConfig{
		Hero:    "batman",
		Villian: "joker",
		Matchups: map[string]string{
			"batman":   "superman",
			"superman": "luther",
		},
		Cities: []string{"gotham", "central", "star"},

		ShootingLocation: "LA",
	}
}

func TestEnvLoadingNoFile(t *testing.T) {
	env := os.Environ()
	os.Clearenv()
	defer func() {
		for _, pair := range env {
			parts := strings.SplitN(pair, "=", 2)
			os.Setenv(parts[0], parts[1])
		}
	}()

	os.Setenv("TEST_VILLIAN", "joker")
	os.Setenv("TEST_HERO", "batman")
	os.Setenv("TEST_MATCHUPS", "batman:superman,superman:luther")
	os.Setenv("TEST_CITIES", "gotham,central,star")
	os.Setenv("TEST_SHOOTING_LOCATION", "LA")

	var results testConfig
	assert.NoError(t, LoadFromEnv("test", "", &results))
	validateConfig(t, exampleConfig(), results)
}

func TestEnvLoadingMissingFile(t *testing.T) {
	err := LoadFromEnv("test", "should-exist.env", &struct{}{})
	assert.Error(t, err)
}

func TestEnvLoadingFromFile(t *testing.T) {
	os.Clearenv()
	data := `
TEST_VILLIAN=joker
TEST_HERO=batman
TEST_MATCHUPS=batman:superman,superman:luther
TEST_CITIES=gotham,central,star
TEST_SHOOTING_LOCATION=LA
`
	filename := writeTestFile(t, "env", []byte(data))
	defer os.Remove(filename)

	var results testConfig
	assert.NoError(t, LoadFromEnv("test", filename, &results))
	validateConfig(t, exampleConfig(), results)
}

func TestFileLoadingNoFile(t *testing.T) {
	var results = testConfig{
		Hero: "flash",
	}
	var expected = testConfig{
		Hero: "flash",
	}
	require.NoError(t, LoadFromFile("", &results))
	validateConfig(t, expected, results)
}

func TestFileLoadJSON(t *testing.T) {
	expected := exampleConfig()

	input := `{
		"hero": "batman",
		"villian": "joker",
		"matchups": {
			"batman": "superman",
			"superman": "luther"
		},
		"cities": ["gotham","central","star"],
		"shooting_location": "LA"
	}`
	filename := writeTestFile(t, "json", []byte(input))
	defer os.Remove(filename)

	var results testConfig
	require.NoError(t, LoadFromFile(filename, &results))
	validateConfig(t, expected, results)
}

func TestFileLoadYAML(t *testing.T) {
	expected := exampleConfig()

	input := `---
hero: batman
villian: joker
matchups:
  batman: superman
  superman: luther
cities:
  - gotham
  - central
  - star
shooting_location: LA`
	filename := writeTestFile(t, "yaml", []byte(input))
	defer os.Remove(filename)

	var results testConfig
	require.NoError(t, LoadFromFile(filename, &results))
	validateConfig(t, expected, results)
}

func TestFileLoadWithSetFields(t *testing.T) {
	expected := testConfig{
		Hero: "wonder woman",
	}
	// serailize it without the villain set
	bytes, err := json.Marshal(&expected)
	require.NoError(t, err)
	filename := writeTestFile(t, "json", bytes)
	defer os.Remove(filename)

	// set a default field
	expected.Villian = "circe"
	expected.Cities = []string{"gotham"}

	var results testConfig
	require.NoError(t, LoadFromFile(filename, &results))

	// this will overwrite ALL the values
	assert.Equal(t, "", results.Villian)
	assert.Equal(t, "wonder woman", results.Hero)
	assert.Len(t, results.Cities, 0)
}

func TestEnvLoadingWithTags(t *testing.T) {
	data := `
NESTED_WITH_TAG=loaded
WITH_TAG=not-loaded
NESTED_WITHOUT_TAG=loaded
NESTED_JAMMEDTOGETHER=loaded
`
	filename := writeTestFile(t, "env", []byte(data))
	defer os.Remove(filename)

	results := struct {
		Nested struct {
			WithTag        string `envconfig:"with_tag"`
			WithoutTag     string `split_words:"true"`
			JammedTogether string
		}
	}{}

	require.NoError(t, LoadFromEnv("", filename, &results))
	assert.Equal(t, "loaded", results.Nested.WithTag)
	assert.Equal(t, "loaded", results.Nested.JammedTogether)
	assert.Equal(t, "loaded", results.Nested.WithoutTag)
}

func writeTestFile(t *testing.T, ext string, data []byte) string {
	f, err := ioutil.TempFile("", "test-*."+ext)
	require.NoError(t, err)

	ioutil.WriteFile(f.Name(), data, 0644)
	require.NoError(t, f.Close())
	return f.Name()
}

func validateConfig(t *testing.T, expected testConfig, results testConfig) {
	assert.Equal(t, expected.Hero, results.Hero)
	assert.Equal(t, expected.Villian, results.Villian)
	assert.Len(t, results.Cities, len(expected.Cities))
	for _, city := range expected.Cities {
		assert.Contains(t, results.Cities, city)
	}

	assert.Len(t, results.Matchups, len(expected.Matchups))
	for k, v := range expected.Matchups {
		assert.Equal(t, v, results.Matchups[k])
	}
	assert.Equal(t, results.ShootingLocation, expected.ShootingLocation)
}
