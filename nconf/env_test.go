package nconf

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvLoadingNoFile(t *testing.T) {
	os.Clearenv()

	os.Setenv("TEST_VILLIAN", "joker")
	os.Setenv("TEST_HERO", "batman")

	out := struct {
		Villian string
		Hero    string
	}{}

	assert.NoError(t, LoadFromEnv("test", "", &out))

	assert.Equal(t, "batman", out.Hero)
	assert.Equal(t, "joker", out.Villian)
}

func TestEnvLoadingMissingFile(t *testing.T) {
	os.Clearenv()
	out := struct {
		Villian string
		Hero    string
	}{}

	err := LoadFromEnv("test", "should-exist.env", &out)
	assert.Error(t, err)
}

func TestEnvLoadingFromFile(t *testing.T) {
	os.Clearenv()

	f, err := ioutil.TempFile("", "env-testing")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	data := "TEST_VILLIAN=joker\nTEST_HERO=batman\n"
	ioutil.WriteFile(f.Name(), []byte(data), 0644)

	out := struct {
		Villian string
		Hero    string
	}{}

	assert.NoError(t, LoadFromEnv("test", f.Name(), &out))

	assert.Equal(t, "batman", out.Hero)
	assert.Equal(t, "joker", out.Villian)
}
