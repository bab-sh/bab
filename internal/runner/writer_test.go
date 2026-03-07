package runner

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/charmbracelet/lipgloss"
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
