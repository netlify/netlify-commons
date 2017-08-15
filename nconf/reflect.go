package nconf

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/spf13/viper"
)

const tagPrefix = "viper"

func recursivelySet(val reflect.Value, prefix string) (bool, error) {
	if val.Kind() != reflect.Ptr {
		return false, errors.New("Must pass a pointer")
	}

	// dereference
	val = reflect.Indirect(val)
	if val.Kind() != reflect.Struct {
		return false, errors.New("must be a reference to a struct")
	}

	// grab the type for this instance
	vType := reflect.TypeOf(val.Interface())

	modified := false
	// go through child fields
	for i := 0; i < val.NumField(); i++ {
		thisField := val.Field(i)
		thisType := vType.Field(i)
		tag := prefix + getTag(thisType)
		modified = modified || viper.IsSet(tag)

		switch thisField.Kind() {
		case reflect.Struct:
			inst := thisField.Addr()
			_, err := recursivelySet(inst, tag+".")
			if err != nil {
				return false, err
			}
		case reflect.Ptr:
			inst := reflect.New(thisField.Type().Elem())
			mod, err := recursivelySet(inst, tag+".")
			if err != nil {
				return false, err
			}
			if mod {
				thisField.Set(inst)
			}
		case reflect.Int:
			fallthrough
		case reflect.Int32:
			fallthrough
		case reflect.Int64:
			// you can only set with an int64 -> int
			configVal := int64(viper.GetInt(tag))
			thisField.SetInt(configVal)
		case reflect.Bool:
			configVal := viper.GetBool(tag)
			thisField.SetBool(configVal)
		case reflect.String:
			configVal := viper.GetString(tag)
			thisField.SetString(configVal)
		case reflect.Slice:
			configVal := viper.GetStringSlice(tag)
			thisField.Set(reflect.ValueOf(configVal))
		default:
			return false, fmt.Errorf("unexpected type detected ~ aborting: %s", thisField.Kind())
		}
	}

	return modified, nil
}

func getTag(field reflect.StructField) string {
	// check if maybe we have a special magic tag
	tag := field.Tag
	if tag != "" {
		for _, prefix := range []string{tagPrefix, "mapstructure", "json"} {
			if v := tag.Get(prefix); v != "" {
				return v
			}
		}
	}

	return field.Name
}
