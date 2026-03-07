package runner

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/bab-sh/bab/internal/theme"
	"github.com/bab-sh/bab/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func sanitizeLine(line string) string {
	if !strings.ContainsRune(line, '\x1b') && !strings.ContainsRune(line, 0x9b) {
		return line
	}
	var b strings.Builder
	b.Grow(len(line))
	var state byte
	p := ansi.NewParser()
	input := line
	for len(input) > 0 {
		seq, _, n, newState := ansi.DecodeSequence(input, state, p)
		if n == 0 {
			break
		}
		state = newState
		if ansi.HasCsiPrefix(seq) {
			if p.Command()&0xff == 'm' {
				b.WriteString(seq)
			}
		} else if !ansi.HasOscPrefix(seq) && !ansi.HasEscPrefix(seq) &&
			!ansi.HasDcsPrefix(seq) && !ansi.HasApcPrefix(seq) {
			b.WriteString(seq)
		}
		input = input[n:]
	}
	return b.String()
}

func cleanLine(line string, strip bool) string {
	if strip {
		return ansi.Strip(line)
	}
	return sanitizeLine(line)
}

type PrefixWriter struct {
	prefix  string
	dest    io.Writer
	mu      *sync.Mutex
	partial []byte
	strip   bool
}

func NewPrefixWriter(label string, padWidth int, color lipgloss.Color, dest io.Writer, mu *sync.Mutex, strip bool) *PrefixWriter {
	style := lipgloss.NewStyle().Foreground(color)
	paddedLabel := fmt.Sprintf("%-*s", padWidth, label)
	prefix := style.Render("["+paddedLabel+"]") + " "
	return &PrefixWriter{
		prefix: prefix,
		dest:   dest,
		mu:     mu,
		strip:  strip,
	}
}

func (pw *PrefixWriter) Write(p []byte) (int, error) {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	total := len(p)
	pw.partial = append(pw.partial, p...)
	data := pw.partial
	pw.partial = nil

	for len(data) > 0 {
		idx := bytes.IndexByte(data, '\n')
		if idx < 0 {
			pw.partial = append(pw.partial, data...)
			break
		}
		line := cleanLine(string(data[:idx]), pw.strip)
		data = data[idx+1:]
		if _, err := fmt.Fprintln(pw.dest, pw.prefix+line); err != nil {
			return total, err
		}
	}

	return total, nil
}

func (pw *PrefixWriter) Flush() error {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	if len(pw.partial) > 0 {
		line := cleanLine(string(pw.partial), pw.strip)
		pw.partial = nil
		_, err := fmt.Fprintln(pw.dest, pw.prefix+line)
		return err
	}
	return nil
}

type LineWriter struct {
	index   int
	program *tea.Program
	mu      sync.Mutex
	partial []byte
	strip   bool
}

func NewLineWriter(index int, program *tea.Program, strip bool) *LineWriter {
	return &LineWriter{
		index:   index,
		program: program,
		strip:   strip,
	}
}

func (lw *LineWriter) Write(p []byte) (int, error) {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	total := len(p)
	lw.partial = append(lw.partial, p...)
	data := lw.partial
	lw.partial = nil

	for len(data) > 0 {
		idx := bytes.IndexByte(data, '\n')
		if idx < 0 {
			lw.partial = append(lw.partial, data...)
			break
		}
		line := cleanLine(string(data[:idx]), lw.strip)
		data = data[idx+1:]
		lw.program.Send(tui.ItemOutputMsg{Index: lw.index, Line: line})
	}

	return total, nil
}

func (lw *LineWriter) Flush() {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	if len(lw.partial) > 0 {
		line := cleanLine(string(lw.partial), lw.strip)
		lw.partial = nil
		lw.program.Send(tui.ItemOutputMsg{Index: lw.index, Line: line})
	}
}

func colorForIndex(index int) lipgloss.Color {
	return theme.ParallelColors[index%len(theme.ParallelColors)]
}
