package cfg

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const _OPTIONAL = "opt"

func GetFromEnv(obj any) error {
	// check obj is pointer
	if reflect.ValueOf(obj).Kind() != reflect.Ptr {
		return errors.New("obj must be a pointer")
	}

	v := reflect.ValueOf(obj).Elem()
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		envTag := field.Tag.Get("env")
		if envTag == "" {
			continue
		}
		var isOptional bool
		if strings.Contains(envTag, ",") {
			envTagSlice := strings.Split(envTag, ",")
			envTag = envTagSlice[0]
			isOptional = (strings.ToLower(envTagSlice[1]) == _OPTIONAL)
		}
		if isOptional {
			fmt.Println("optional", envTag)
		}

		envValue := os.Getenv(envTag)
		if envValue == "" && isOptional {
			continue
		}
		if envValue == "" {
			return errors.New("environmental variable " + envTag + " must not be blank")
		}

		fieldValue := v.Field(i)
		switch fieldValue.Kind() {
		case reflect.String:
			fieldValue.SetString(envValue)
		case reflect.Int:
			intValue, err := strconv.Atoi(envValue)
			if err != nil {
				return errors.New("environmental variable " + envTag + " must be an integer")
			}
			fieldValue.SetInt(int64(intValue))
		case reflect.Bool:
			boolValue, err := strconv.ParseBool(envValue)
			if err != nil {
				return errors.New("environmental variable " + envTag + " must be a boolean")
			}
			fieldValue.SetBool(boolValue)
		case reflect.TypeOf(time.Duration(0)).Kind():
			durationValue, err := time.ParseDuration(envValue)
			if err != nil {
				return errors.New("environmental variable " + envTag + " must be a duration")
			}
			fieldValue.Set(reflect.ValueOf(durationValue))
		default:
			return errors.New("unsupported type: " + fieldValue.Kind().String())
		}
	}
	return nil
}
