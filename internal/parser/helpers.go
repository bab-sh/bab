package parser

import "fmt"

func safeStringCast(v interface{}) (string, error) {
	if v == nil {
		return "", fmt.Errorf("cannot cast nil to string")
	}

	if s, ok := v.(string); ok {
		return s, nil
	}

	return fmt.Sprint(v), nil
}

func safeMapCast(v interface{}) (map[string]interface{}, bool) {
	if v == nil {
		return nil, false
	}

	m, ok := v.(map[string]interface{})
	return m, ok
}

func safeSliceCast(v interface{}) ([]interface{}, bool) {
	if v == nil {
		return nil, false
	}

	s, ok := v.([]interface{})
	return s, ok
}

func safeStringSliceCast(v interface{}) ([]string, bool) {
	if v == nil {
		return nil, false
	}

	slice, ok := v.([]interface{})
	if !ok {
		return nil, false
	}

	result := make([]string, 0, len(slice))
	for _, item := range slice {
		str, err := safeStringCast(item)
		if err != nil {
			return nil, false
		}
		result = append(result, str)
	}
	return result, true
}
