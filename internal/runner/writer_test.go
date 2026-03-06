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
	pw := NewPrefixWriter("test", 4, lipgloss.Color("0"), &buf, mu)

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
	pw := NewPrefixWriter("app", 3, lipgloss.Color("0"), &buf, mu)

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
	pw := NewPrefixWriter("x", 1, lipgloss.Color("0"), &buf, mu)

	_, _ = pw.Write([]byte("hel"))
	if buf.Len() != 0 {
		t.Error("expected no output for partial line")
	}

	_, _ = pw.Write([]byte("lo\n"))
	if !strings.Contains(buf.String(), "hello") {
		t.Errorf("expected flushed line to contain 'hello', got %q", buf.String())
	}
}

func TestPrefixWriterFlush(t *testing.T) {
	var buf bytes.Buffer
	mu := &sync.Mutex{}
	pw := NewPrefixWriter("x", 1, lipgloss.Color("0"), &buf, mu)

	_, _ = pw.Write([]byte("partial"))
	_ = pw.Flush()
	if !strings.Contains(buf.String(), "partial") {
		t.Errorf("Flush should write partial line, got %q", buf.String())
	}
}
