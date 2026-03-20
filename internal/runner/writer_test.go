package runner

import (
	"bytes"
	"image/color"
	"strings"
	"sync"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/bab-sh/bab/internal/theme"
)

func TestPrefixWriterSingleLine(t *testing.T) {
	var buf bytes.Buffer
	mu := &sync.Mutex{}
	pw := NewPrefixWriter("test", 4, lipgloss.Color("0"), &buf, mu, false)

	_, err := pw.Write([]byte("hello world\n"))
	if err != nil {
		t.Fatalf("Write error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "test") {
		t.Errorf("expected output to contain label 'test', got %q", output)
	}
	if !strings.Contains(output, "hello world") {
		t.Errorf("expected output to contain 'hello world', got %q", output)
	}
}

func TestPrefixWriterMultipleLines(t *testing.T) {
	var buf bytes.Buffer
	mu := &sync.Mutex{}
	pw := NewPrefixWriter("app", 3, lipgloss.Color("0"), &buf, mu, false)

	_, _ = pw.Write([]byte("line 1\nline 2\n"))

	output := buf.String()
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), output)
	}
	for _, line := range lines {
		if !strings.Contains(line, "app") {
			t.Errorf("each line should have prefix, got %q", line)
		}
	}
}

func TestPrefixWriterPartialLine(t *testing.T) {
	var buf bytes.Buffer
	mu := &sync.Mutex{}
	pw := NewPrefixWriter("x", 1, lipgloss.Color("0"), &buf, mu, false)

	_, _ = pw.Write([]byte("hel"))
	if buf.Len() != 0 {
		t.Error("expected no output for partial line")
	}

	_, _ = pw.Write([]byte("lo\n"))
	if !strings.Contains(buf.String(), "hello") {
		t.Errorf("expected flushed line to contain 'hello', got %q", buf.String())
	}
}

func TestSanitizeLineKeepsSGR(t *testing.T) {
	input := "\x1b[31mred text\x1b[0m"
	got := sanitizeLine(input)
	if got != input {
		t.Errorf("expected SGR to be kept, got %q", got)
	}
}

func TestSanitizeLineStripsDangerous(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"cursor move", "\x1b[2Jhello", "hello"},
		{"clear screen", "\x1b[Hhello", "hello"},
		{"alt screen", "\x1b[?1049hhello", "hello"},
		{"OSC title", "\x1b]0;title\x07hello", "hello"},
		{"plain text", "no escapes here", "no escapes here"},
		{"SGR + dangerous", "\x1b[31mred\x1b[2J gone\x1b[0m", "\x1b[31mred gone\x1b[0m"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeLine(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeLine(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCleanLineStripAll(t *testing.T) {
	input := "\x1b[31mcolored\x1b[0m"
	got := cleanLine(input, true)
	if got != "colored" {
		t.Errorf("cleanLine(strip=true) should remove all ANSI, got %q", got)
	}
}

func TestCleanLineKeepSGR(t *testing.T) {
	input := "\x1b[31mcolored\x1b[0m"
	got := cleanLine(input, false)
	if got != input {
		t.Errorf("cleanLine(strip=false) should keep SGR, got %q", got)
	}
}

func TestPrefixWriterFlush(t *testing.T) {
	var buf bytes.Buffer
	mu := &sync.Mutex{}
	pw := NewPrefixWriter("x", 1, lipgloss.Color("0"), &buf, mu, false)

	_, _ = pw.Write([]byte("partial"))
	_ = pw.Flush()
	if !strings.Contains(buf.String(), "partial") {
		t.Errorf("Flush should write partial line, got %q", buf.String())
	}
}

func TestLineBuffer(t *testing.T) {
	tests := []struct {
		name     string
		writes   []string
		flush    bool
		strip    bool
		wantEmit []string
	}{
		{
			name:     "single complete line",
			writes:   []string{"hello\n"},
			wantEmit: []string{"hello"},
		},
		{
			name:     "multiple lines in one write",
			writes:   []string{"line1\nline2\n"},
			wantEmit: []string{"line1", "line2"},
		},
		{
			name:     "partial line buffered",
			writes:   []string{"hel"},
			wantEmit: nil,
		},
		{
			name:     "partial then complete",
			writes:   []string{"hel", "lo\n"},
			wantEmit: []string{"hello"},
		},
		{
			name:     "flush with buffered content",
			writes:   []string{"partial"},
			flush:    true,
			wantEmit: []string{"partial"},
		},
		{
			name:     "flush with empty buffer",
			writes:   []string{"done\n"},
			flush:    true,
			wantEmit: []string{"done"},
		},
		{
			name:     "strip mode removes ANSI",
			writes:   []string{"\x1b[31mred\x1b[0m\n"},
			strip:    true,
			wantEmit: []string{"red"},
		},
		{
			name:     "non-strip keeps SGR",
			writes:   []string{"\x1b[31mred\x1b[0m\n"},
			strip:    false,
			wantEmit: []string{"\x1b[31mred\x1b[0m"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lb := &lineBuffer{strip: tt.strip}
			var emitted []string
			emit := func(line string) error {
				emitted = append(emitted, line)
				return nil
			}
			for _, w := range tt.writes {
				if _, err := lb.process([]byte(w), emit); err != nil {
					t.Fatalf("process() error: %v", err)
				}
			}
			if tt.flush {
				if err := lb.flush(emit); err != nil {
					t.Fatalf("flush() error: %v", err)
				}
			}
			if len(emitted) != len(tt.wantEmit) {
				t.Fatalf("emitted %d lines, want %d: %v", len(emitted), len(tt.wantEmit), emitted)
			}
			for i, want := range tt.wantEmit {
				if emitted[i] != want {
					t.Errorf("emitted[%d] = %q, want %q", i, emitted[i], want)
				}
			}
		})
	}
}

func colorsEqual(a, b color.Color) bool {
	r1, g1, b1, a1 := a.RGBA()
	r2, g2, b2, a2 := b.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}

func TestColorForPath(t *testing.T) {
	tests := []struct {
		name string
		path []int
		want color.Color
	}{
		{"empty path", nil, theme.ParallelBaseColors[0]},
		{"single index 0", []int{0}, theme.ParallelBaseColors[0]},
		{"single index 1", []int{1}, theme.ParallelBaseColors[1]},
		{"wraps around", []int{len(theme.ParallelBaseColors)}, theme.ParallelBaseColors[0]},
		{"nested path is dimmed", []int{0, 0}, dimColor(theme.ParallelBaseColors[0], 1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := colorForPath(tt.path)
			if !colorsEqual(got, tt.want) {
				t.Errorf("colorForPath(%v) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestDimColor(t *testing.T) {
	tests := []struct {
		name  string
		color color.Color
		steps int
		check func(t *testing.T, result color.Color)
	}{
		{
			name:  "color code below 16",
			color: lipgloss.Color("10"),
			steps: 1,
			check: func(t *testing.T, result color.Color) {
				if !colorsEqual(result, lipgloss.Color("10")) {
					t.Errorf("expected unchanged, got %v", result)
				}
			},
		},
		{
			name:  "color code above 231",
			color: lipgloss.Color("240"),
			steps: 1,
			check: func(t *testing.T, result color.Color) {
				if !colorsEqual(result, lipgloss.Color("240")) {
					t.Errorf("expected unchanged, got %v", result)
				}
			},
		},
		{
			name:  "valid 256-color dimmed",
			color: lipgloss.Color("196"),
			steps: 1,
			check: func(t *testing.T, result color.Color) {
				if colorsEqual(result, lipgloss.Color("196")) {
					t.Error("expected color to be dimmed")
				}
			},
		},
		{
			name:  "multiple steps progressively dimmer",
			color: lipgloss.Color("196"),
			steps: 2,
			check: func(t *testing.T, result color.Color) {
				oneStep := dimColor(lipgloss.Color("196"), 1)
				if colorsEqual(result, oneStep) {
					t.Error("two steps should be dimmer than one step")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dimColor(tt.color, tt.steps)
			tt.check(t, result)
		})
	}
}
