package parser

import (
	"testing"
)

func TestSafeStringCast(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    string
		wantErr bool
	}{
		{
			name:    "string value",
			input:   "hello",
			want:    "hello",
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: false,
		},
		{
			name:    "integer value",
			input:   123,
			want:    "123",
			wantErr: false,
		},
		{
			name:    "float value",
			input:   45.67,
			want:    "45.67",
			wantErr: false,
		},
		{
			name:    "boolean true",
			input:   true,
			want:    "true",
			wantErr: false,
		},
		{
			name:    "boolean false",
			input:   false,
			want:    "false",
			wantErr: false,
		},
		{
			name:    "nil value",
			input:   nil,
			want:    "",
			wantErr: true,
		},
		{
			name:    "slice value",
			input:   []string{"a", "b"},
			want:    "[a b]",
			wantErr: false,
		},
		{
			name:    "map value",
			input:   map[string]string{"key": "value"},
			want:    "map[key:value]",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := safeStringCast(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Error("safeStringCast() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("safeStringCast() unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("safeStringCast() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSafeMapCast(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		wantOk   bool
		validate func(t *testing.T, got map[string]interface{})
	}{
		{
			name: "valid map",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": 123,
			},
			wantOk: true,
			validate: func(t *testing.T, got map[string]interface{}) {
				if len(got) != 2 {
					t.Errorf("expected map length 2, got %d", len(got))
				}
				if got["key1"] != "value1" {
					t.Errorf("expected key1=value1, got %v", got["key1"])
				}
				if got["key2"] != 123 {
					t.Errorf("expected key2=123, got %v", got["key2"])
				}
			},
		},
		{
			name:   "empty map",
			input:  map[string]interface{}{},
			wantOk: true,
			validate: func(t *testing.T, got map[string]interface{}) {
				if len(got) != 0 {
					t.Errorf("expected empty map, got %d items", len(got))
				}
			},
		},
		{
			name:   "nil value",
			input:  nil,
			wantOk: false,
		},
		{
			name:   "string value",
			input:  "not a map",
			wantOk: false,
		},
		{
			name:   "integer value",
			input:  123,
			wantOk: false,
		},
		{
			name:   "slice value",
			input:  []interface{}{"a", "b"},
			wantOk: false,
		},
		{
			name:   "boolean value",
			input:  true,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := safeMapCast(tt.input)

			if ok != tt.wantOk {
				t.Errorf("safeMapCast() ok = %v, want %v", ok, tt.wantOk)
				return
			}

			if !tt.wantOk {
				if got != nil {
					t.Error("safeMapCast() expected nil map when ok=false")
				}
				return
			}

			if tt.validate != nil {
				tt.validate(t, got)
			}
		})
	}
}

func TestSafeSliceCast(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		wantOk   bool
		validate func(t *testing.T, got []interface{})
	}{
		{
			name:   "valid slice",
			input:  []interface{}{"a", "b", "c"},
			wantOk: true,
			validate: func(t *testing.T, got []interface{}) {
				if len(got) != 3 {
					t.Errorf("expected slice length 3, got %d", len(got))
				}
				if got[0] != "a" {
					t.Errorf("expected got[0]=a, got %v", got[0])
				}
			},
		},
		{
			name:   "empty slice",
			input:  []interface{}{},
			wantOk: true,
			validate: func(t *testing.T, got []interface{}) {
				if len(got) != 0 {
					t.Errorf("expected empty slice, got %d items", len(got))
				}
			},
		},
		{
			name:   "slice with mixed types",
			input:  []interface{}{"string", 123, true, nil},
			wantOk: true,
			validate: func(t *testing.T, got []interface{}) {
				if len(got) != 4 {
					t.Errorf("expected slice length 4, got %d", len(got))
				}
			},
		},
		{
			name:   "nil value",
			input:  nil,
			wantOk: false,
		},
		{
			name:   "string value",
			input:  "not a slice",
			wantOk: false,
		},
		{
			name:   "integer value",
			input:  123,
			wantOk: false,
		},
		{
			name:   "map value",
			input:  map[string]interface{}{"key": "value"},
			wantOk: false,
		},
		{
			name:   "boolean value",
			input:  false,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := safeSliceCast(tt.input)

			if ok != tt.wantOk {
				t.Errorf("safeSliceCast() ok = %v, want %v", ok, tt.wantOk)
				return
			}

			if !tt.wantOk {
				if got != nil {
					t.Error("safeSliceCast() expected nil slice when ok=false")
				}
				return
			}

			if tt.validate != nil {
				tt.validate(t, got)
			}
		})
	}
}
