package nconf

import (
	"reflect"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestSimpleValues(t *testing.T) {
	c := struct {
		Simple string `json:"simple"`
	}{}

	viper.SetDefault("simple", "i am a simple string")

	mod, err := recursivelySet(reflect.ValueOf(&c), "")
	assert.Nil(t, err)
	assert.True(t, mod)
	assert.Equal(t, "i am a simple string", c.Simple)
}

func TestNestedValues(t *testing.T) {
	c := struct {
		Simple string `json:"simple"`
		Nested struct {
			BoolVal   bool   `json:"bool"`
			StringVal string `json:"string"`
			NumberVal int    `json:"number"`
		} `json:"nested"`
		NestedPtr *struct {
			String string `json:"string"`
		} `json:"pointer"`
		MissingPtr *struct {
			String string `json:"string"`
		} `json:"missing"`
		Slice []string `mapstructure:"slice"`
	}{}

	viper.SetDefault("simple", "simple")
	viper.SetDefault("nested.bool", true)
	viper.SetDefault("nested.string", "i am a simple string")
	viper.SetDefault("nested.number", 4)
	viper.SetDefault("slice", []string{"something", "good"})
	viper.SetDefault("pointer.string", "i am a string too")

	mod, err := recursivelySet(reflect.ValueOf(&c), "")
	assert.Nil(t, err)
	assert.True(t, mod)
	assert.Equal(t, "simple", c.Simple)
	assert.Equal(t, 4, c.Nested.NumberVal)
	assert.Equal(t, "i am a simple string", c.Nested.StringVal)
	assert.Equal(t, true, c.Nested.BoolVal)
	assert.NotNil(t, c.NestedPtr)
	assert.Equal(t, "i am a string too", c.NestedPtr.String)
	assert.Nil(t, c.MissingPtr)
	assert.Equal(t, c.Slice, []string{"something", "good"})
	assert.Len(t, c.Slice, 2)
}
