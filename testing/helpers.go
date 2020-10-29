package testing

import "reflect"

func ToMap(m map[string]interface{}) map[string]interface{} {
	cp := make(map[string]interface{})

	for k, v := range m {
		switch cv := v.(type) {
		case map[string]interface{}:
			cp[k] = ToMap(cv)
		default:
			cp[k] = TryToMap(cv)
		}
	}

	return cp
}

func TryToMap(input interface{}) interface{} {
	t := reflect.ValueOf(input)

	if t.Kind() == reflect.Struct {
		sm := make(map[string]interface{})

		for x := 0; x < t.NumField(); x++ {
			fieldValue := t.Field(x)
			field := t.Type().Field(x)

			name := field.Name
			tag := field.Tag.Get("json")

			if tag != "" {
				name = tag
			}

			sm[name] = TryToMap(fieldValue.Interface())
		}

		return sm
	}

	return input
}
