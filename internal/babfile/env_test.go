package babfile

import (
	"sort"
	"testing"
)

func TestMergeEnvMaps(t *testing.T) {
	tests := []struct {
		name string
		maps []map[string]string
		want map[string]string
	}{
		{
			name: "empty maps",
			maps: []map[string]string{},
			want: map[string]string{},
		},
		{
			name: "single map",
			maps: []map[string]string{
				{"FOO": "bar", "BAZ": "qux"},
			},
			want: map[string]string{"FOO": "bar", "BAZ": "qux"},
		},
		{
			name: "merge two maps",
			maps: []map[string]string{
				{"FOO": "bar"},
				{"BAZ": "qux"},
			},
			want: map[string]string{"FOO": "bar", "BAZ": "qux"},
		},
		{
			name: "later overrides earlier",
			maps: []map[string]string{
				{"FOO": "original"},
				{"FOO": "override"},
			},
			want: map[string]string{"FOO": "override"},
		},
		{
			name: "three level merge",
			maps: []map[string]string{
				{"A": "global", "B": "global"},
				{"B": "task", "C": "task"},
				{"C": "cmd"},
			},
			want: map[string]string{"A": "global", "B": "task", "C": "cmd"},
		},
		{
			name: "nil maps are skipped",
			maps: []map[string]string{
				{"FOO": "bar"},
				nil,
				{"BAZ": "qux"},
			},
			want: map[string]string{"FOO": "bar", "BAZ": "qux"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeEnvMaps(tt.maps...)
			if len(got) != len(tt.want) {
				t.Errorf("MergeEnvMaps() len = %d, want %d", len(got), len(tt.want))
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("MergeEnvMaps()[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}

func TestMergeEnv(t *testing.T) {
	tests := []struct {
		name string
		maps []map[string]string
		want []string
	}{
		{
			name: "empty maps",
			maps: []map[string]string{},
			want: []string{},
		},
		{
			name: "single map",
			maps: []map[string]string{
				{"FOO": "bar"},
			},
			want: []string{"FOO=bar"},
		},
		{
			name: "merged maps",
			maps: []map[string]string{
				{"FOO": "original"},
				{"FOO": "override", "BAR": "baz"},
			},
			want: []string{"BAR=baz", "FOO=override"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeEnv(tt.maps...)
			sort.Strings(got)
			sort.Strings(tt.want)
			if len(got) != len(tt.want) {
				t.Errorf("MergeEnv() len = %d, want %d", len(got), len(tt.want))
			}
			for i, v := range tt.want {
				if got[i] != v {
					t.Errorf("MergeEnv()[%d] = %q, want %q", i, got[i], v)
				}
			}
		})
	}
}

func TestEnvSchema(t *testing.T) {
	schema := EnvSchema()
	if schema.Type != "object" {
		t.Errorf("EnvSchema().Type = %q, want %q", schema.Type, "object")
	}
	if schema.AdditionalProperties == nil {
		t.Error("EnvSchema().AdditionalProperties should not be nil")
	}
	if schema.AdditionalProperties.Type != "string" {
		t.Errorf("EnvSchema().AdditionalProperties.Type = %q, want %q", schema.AdditionalProperties.Type, "string")
	}
}
