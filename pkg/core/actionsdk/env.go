package actionsdk

import (
	"os"
	"reflect"
	"strconv"
	"strings"
)

const envVarTag = "env"

// PopulateFromEnv updates struct fields using values from a provided env map.
func PopulateFromEnv(data interface{}, env map[string]string) {
	if data == nil || env == nil {
		return
	}

	v := reflect.ValueOf(data)

	// Ensure the input is a pointer to a struct
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return
	}

	val := v.Elem()
	t := val.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get(envVarTag)
		if tag == "" {
			continue
		}

		envVal, exists := env[tag]
		if !exists || strings.TrimSpace(envVal) == "" || strings.EqualFold(envVal, "null") {
			continue
		}

		fieldVal := val.Field(i)
		if !fieldVal.CanSet() {
			continue
		}

		switch fieldVal.Kind() {
		case reflect.String:
			fieldVal.SetString(envVal)

		case reflect.Int, reflect.Int64:
			if parsedVal, err := strconv.ParseInt(envVal, 10, 64); err == nil {
				fieldVal.SetInt(parsedVal)
			}

		case reflect.Float64:
			if parsedVal, err := strconv.ParseFloat(envVal, 64); err == nil {
				fieldVal.SetFloat(parsedVal)
			}

		case reflect.Bool:
			if parsedVal, err := strconv.ParseBool(envVal); err == nil {
				fieldVal.SetBool(parsedVal)
			}

		case reflect.Slice:
			if field.Type.Elem().Kind() == reflect.String {
				fieldVal.Set(reflect.ValueOf(strings.Split(envVal, ",")))
			}

		default:
			// Unsupported type, skipping
		}
	}
}

// EnvMap retrieves all environment variables and returns them as a map.
func EnvMap() map[string]string {
	envMap := make(map[string]string)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}
	return envMap
}
