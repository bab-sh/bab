package parser

import "fmt"

func normalizeMap(v interface{}) interface{} {
	switch val := v.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]interface{}, len(val))
		for k, v := range val {
			m[fmt.Sprint(k)] = normalizeMap(v)
		}
		return m
	case map[string]interface{}:
		m := make(map[string]interface{}, len(val))
		for k, v := range val {
			m[k] = normalizeMap(v)
		}
		return m
	case []interface{}:
		for i, item := range val {
			val[i] = normalizeMap(item)
		}
		return val
	default:
		return v
	}
}
