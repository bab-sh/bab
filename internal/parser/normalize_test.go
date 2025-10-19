package parser

import (
	"reflect"
	"testing"
)

func TestNormalizeMap(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  interface{}
	}{
		{
			name: "map[interface{}]interface{} to map[string]interface{}",
			input: map[interface{}]interface{}{
				"key1": "value1",
				"key2": 123,
			},
			want: map[string]interface{}{
				"key1": "value1",
				"key2": 123,
			},
		},
		{
			name: "already normalized map[string]interface{}",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": 456,
			},
			want: map[string]interface{}{
				"key1": "value1",
				"key2": 456,
			},
		},
		{
			name: "nested map[interface{}]interface{}",
			input: map[interface{}]interface{}{
				"parent": map[interface{}]interface{}{
					"child": "value",
				},
			},
			want: map[string]interface{}{
				"parent": map[string]interface{}{
					"child": "value",
				},
			},
		},
		{
			name: "nested map[string]interface{}",
			input: map[string]interface{}{
				"parent": map[string]interface{}{
					"child": "value",
				},
			},
			want: map[string]interface{}{
				"parent": map[string]interface{}{
					"child": "value",
				},
			},
		},
		{
			name: "slice with maps",
			input: []interface{}{
				map[interface{}]interface{}{"key": "value"},
				"string",
				123,
			},
			want: []interface{}{
				map[string]interface{}{"key": "value"},
				"string",
				123,
			},
		},
		{
			name:  "string value unchanged",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "integer value unchanged",
			input: 789,
			want:  789,
		},
		{
			name:  "boolean value unchanged",
			input: true,
			want:  true,
		},
		{
			name:  "nil value unchanged",
			input: nil,
			want:  nil,
		},
		{
			name: "complex nested structure",
			input: map[interface{}]interface{}{
				"ci": map[interface{}]interface{}{
					"test": map[interface{}]interface{}{
						"run": []interface{}{
							"go test",
							"go vet",
						},
						"desc": "Run tests",
					},
				},
			},
			want: map[string]interface{}{
				"ci": map[string]interface{}{
					"test": map[string]interface{}{
						"run": []interface{}{
							"go test",
							"go vet",
						},
						"desc": "Run tests",
					},
				},
			},
		},
		{
			name: "map with integer keys",
			input: map[interface{}]interface{}{
				123: "value1",
				456: "value2",
			},
			want: map[string]interface{}{
				"123": "value1",
				"456": "value2",
			},
		},
		{
			name: "empty map[interface{}]interface{}",
			input: map[interface{}]interface{}{},
			want:  map[string]interface{}{},
		},
		{
			name:  "empty map[string]interface{}",
			input: map[string]interface{}{},
			want:  map[string]interface{}{},
		},
		{
			name:  "empty slice",
			input: []interface{}{},
			want:  []interface{}{},
		},
		{
			name: "deeply nested maps and slices",
			input: map[interface{}]interface{}{
				"level1": map[interface{}]interface{}{
					"level2": []interface{}{
						map[interface{}]interface{}{
							"level3": "value",
						},
					},
				},
			},
			want: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": []interface{}{
						map[string]interface{}{
							"level3": "value",
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeMap(tt.input)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("normalizeMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeMapTypes(t *testing.T) {
	t.Run("output is always map[string]interface{} for map input", func(t *testing.T) {
		input := map[interface{}]interface{}{
			"key": "value",
		}
		result := normalizeMap(input)

		_, ok := result.(map[string]interface{})
		if !ok {
			t.Errorf("normalizeMap() returned %T, expected map[string]interface{}", result)
		}
	})

	t.Run("preserves slice type", func(t *testing.T) {
		input := []interface{}{"a", "b", "c"}
		result := normalizeMap(input)

		_, ok := result.([]interface{})
		if !ok {
			t.Errorf("normalizeMap() returned %T, expected []interface{}", result)
		}
	})

	t.Run("preserves string type", func(t *testing.T) {
		input := "hello"
		result := normalizeMap(input)

		_, ok := result.(string)
		if !ok {
			t.Errorf("normalizeMap() returned %T, expected string", result)
		}
	})
}
