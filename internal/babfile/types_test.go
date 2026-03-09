package babfile

import (
	"sort"
	"testing"
)

func TestPlatformValid(t *testing.T) {
	tests := []struct {
		platform Platform
		want     bool
	}{
		{PlatformLinux, true},
		{PlatformDarwin, true},
		{PlatformWindows, true},
		{Platform("freebsd"), false},
		{Platform(""), false},
	}
	for _, tt := range tests {
		t.Run(string(tt.platform), func(t *testing.T) {
			if got := tt.platform.Valid(); got != tt.want {
				t.Errorf("Platform(%q).Valid() = %v, want %v", tt.platform, got, tt.want)
			}
		})
	}
}

func TestPlatformString(t *testing.T) {
	tests := []struct {
		platform Platform
		want     string
	}{
		{PlatformLinux, "linux"},
		{PlatformDarwin, "darwin"},
		{PlatformWindows, "windows"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.platform.String(); got != tt.want {
				t.Errorf("Platform(%q).String() = %q, want %q", tt.platform, got, tt.want)
			}
		})
	}
}

func TestParallelModeValid(t *testing.T) {
	tests := []struct {
		mode ParallelMode
		want bool
	}{
		{ParallelInterleaved, true},
		{ParallelGrouped, true},
		{ParallelMode("unknown"), false},
		{ParallelMode(""), false},
	}
	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			if got := tt.mode.Valid(); got != tt.want {
				t.Errorf("ParallelMode(%q).Valid() = %v, want %v", tt.mode, got, tt.want)
			}
		})
	}
}

func TestPromptTypeValid(t *testing.T) {
	tests := []struct {
		pt   PromptType
		want bool
	}{
		{PromptTypeConfirm, true},
		{PromptTypeInput, true},
		{PromptTypeSelect, true},
		{PromptTypeMultiselect, true},
		{PromptTypePassword, true},
		{PromptTypeNumber, true},
		{PromptType("invalid"), false},
		{PromptType(""), false},
	}
	for _, tt := range tests {
		t.Run(string(tt.pt), func(t *testing.T) {
			if got := tt.pt.Valid(); got != tt.want {
				t.Errorf("PromptType(%q).Valid() = %v, want %v", tt.pt, got, tt.want)
			}
		})
	}
}

func TestPromptTypeRequiresOptions(t *testing.T) {
	tests := []struct {
		pt   PromptType
		want bool
	}{
		{PromptTypeSelect, true},
		{PromptTypeMultiselect, true},
		{PromptTypeConfirm, false},
		{PromptTypeInput, false},
		{PromptTypePassword, false},
		{PromptTypeNumber, false},
	}
	for _, tt := range tests {
		t.Run(string(tt.pt), func(t *testing.T) {
			if got := tt.pt.RequiresOptions(); got != tt.want {
				t.Errorf("PromptType(%q).RequiresOptions() = %v, want %v", tt.pt, got, tt.want)
			}
		})
	}
}

func TestLogLevelValid(t *testing.T) {
	tests := []struct {
		level LogLevel
		want  bool
	}{
		{LogLevelDebug, true},
		{LogLevelInfo, true},
		{LogLevelWarn, true},
		{LogLevelError, true},
		{LogLevel("trace"), false},
		{LogLevel(""), false},
	}
	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			if got := tt.level.Valid(); got != tt.want {
				t.Errorf("LogLevel(%q).Valid() = %v, want %v", tt.level, got, tt.want)
			}
		})
	}
}

func TestMergeVarMaps(t *testing.T) {
	tests := []struct {
		name string
		maps []VarMap
		want VarMap
	}{
		{
			name: "empty",
			maps: []VarMap{},
			want: VarMap{},
		},
		{
			name: "single map",
			maps: []VarMap{{"FOO": "bar", "BAZ": "qux"}},
			want: VarMap{"FOO": "bar", "BAZ": "qux"},
		},
		{
			name: "merge two maps",
			maps: []VarMap{{"FOO": "bar"}, {"BAZ": "qux"}},
			want: VarMap{"FOO": "bar", "BAZ": "qux"},
		},
		{
			name: "later overrides earlier",
			maps: []VarMap{{"FOO": "original"}, {"FOO": "override"}},
			want: VarMap{"FOO": "override"},
		},
		{
			name: "nil maps are skipped",
			maps: []VarMap{{"FOO": "bar"}, nil, {"BAZ": "qux"}},
			want: VarMap{"FOO": "bar", "BAZ": "qux"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeVarMaps(tt.maps...)
			if len(got) != len(tt.want) {
				t.Errorf("MergeVarMaps() len = %d, want %d", len(got), len(tt.want))
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("MergeVarMaps()[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}

func TestMatchesPlatform(t *testing.T) {
	tests := []struct {
		name      string
		platforms []Platform
		platform  string
		want      bool
	}{
		{"empty slice matches all", nil, "linux", true},
		{"match", []Platform{PlatformLinux}, "linux", true},
		{"no match", []Platform{PlatformLinux}, "darwin", false},
		{"multi with one match", []Platform{PlatformLinux, PlatformDarwin}, "darwin", true},
		{"multi no match", []Platform{PlatformLinux, PlatformWindows}, "darwin", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesPlatform(tt.platforms, tt.platform); got != tt.want {
				t.Errorf("matchesPlatform(%v, %q) = %v, want %v", tt.platforms, tt.platform, got, tt.want)
			}
		})
	}
}

func TestTaskGetAllAliases(t *testing.T) {
	tests := []struct {
		name string
		task *Task
		want []string
	}{
		{"nil task", nil, nil},
		{"no aliases", &Task{}, []string{}},
		{"alias only", &Task{Alias: "a"}, []string{"a"}},
		{"aliases only", &Task{Aliases: []string{"b", "c"}}, []string{"b", "c"}},
		{"both", &Task{Alias: "a", Aliases: []string{"b", "c"}}, []string{"a", "b", "c"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.task.GetAllAliases()
			if tt.want == nil {
				if got != nil {
					t.Errorf("GetAllAliases() = %v, want nil", got)
				}
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("GetAllAliases() len = %d, want %d", len(got), len(tt.want))
			}
			for i, v := range tt.want {
				if got[i] != v {
					t.Errorf("GetAllAliases()[%d] = %q, want %q", i, got[i], v)
				}
			}
		})
	}
}

func TestTaskMapHas(t *testing.T) {
	tm := TaskMap{"foo": &Task{Name: "foo"}}
	if !tm.Has("foo") {
		t.Error("Has(foo) should be true")
	}
	if tm.Has("bar") {
		t.Error("Has(bar) should be false")
	}
	empty := TaskMap{}
	if empty.Has("anything") {
		t.Error("empty map Has() should be false")
	}
}

func TestTaskMapNames(t *testing.T) {
	tm := TaskMap{
		"alpha": &Task{Name: "alpha"},
		"beta":  &Task{Name: "beta"},
	}
	names := tm.Names()
	sort.Strings(names)
	if len(names) != 2 || names[0] != "alpha" || names[1] != "beta" {
		t.Errorf("Names() = %v, want [alpha beta]", names)
	}

	empty := TaskMap{}
	if len(empty.Names()) != 0 {
		t.Error("empty map Names() should be empty")
	}
}

func TestParallelRunUseColor(t *testing.T) {
	trueVal := true
	falseVal := false
	tests := []struct {
		name  string
		color *bool
		want  bool
	}{
		{"nil defaults to true", nil, true},
		{"explicit true", &trueVal, true},
		{"explicit false", &falseVal, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := ParallelRun{Color: tt.color}
			if got := pr.UseColor(); got != tt.want {
				t.Errorf("UseColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParallelRunItemLabel(t *testing.T) {
	pr := ParallelRun{
		Items: []RunItem{
			TaskRun{Task: "build"},
			CommandRun{Cmd: "echo hello world"},
			LogRun{Log: "this is a log message"},
		},
		Labels: []string{"custom", "", ""},
	}

	tests := []struct {
		name  string
		index int
		want  string
	}{
		{"explicit label", 0, "custom"},
		{"CommandRun fallback", 1, "echo hello world"},
		{"LogRun truncation", 2, "this is a log messag"},
		{"out of range", 5, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pr.ItemLabel(tt.index); got != tt.want {
				t.Errorf("ItemLabel(%d) = %q, want %q", tt.index, got, tt.want)
			}
		})
	}
}

func TestParallelRunItemLabelTaskRun(t *testing.T) {
	pr := ParallelRun{
		Items:  []RunItem{TaskRun{Task: "deploy"}},
		Labels: []string{""},
	}
	if got := pr.ItemLabel(0); got != "deploy" {
		t.Errorf("ItemLabel(TaskRun) = %q, want %q", got, "deploy")
	}
}

func TestTruncateRunes(t *testing.T) {
	tests := []struct {
		name string
		s    string
		max  int
		want string
	}{
		{"short", "hi", 10, "hi"},
		{"exact max", "hello", 5, "hello"},
		{"over max", "hello world", 5, "hello"},
		{"multi-byte", "héllo wörld", 5, "héllo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncateRunes(tt.s, tt.max); got != tt.want {
				t.Errorf("truncateRunes(%q, %d) = %q, want %q", tt.s, tt.max, got, tt.want)
			}
		})
	}
}
