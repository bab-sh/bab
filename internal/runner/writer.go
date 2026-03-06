package runner

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/bab-sh/bab/internal/theme"
	"github.com/bab-sh/bab/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PrefixWriter struct {
	prefix  string
	dest    io.Writer
	mu      *sync.Mutex
	partial []byte
}

func NewPrefixWriter(label string, padWidth int, color lipgloss.Color, dest io.Writer, mu *sync.Mutex) *PrefixWriter {
	style := lipgloss.NewStyle().Foreground(color)
	paddedLabel := fmt.Sprintf("%-*s", padWidth, label)
	prefix := style.Render("["+paddedLabel+"]") + " "
	return &PrefixWriter{
		prefix: prefix,
		dest:   dest,
		mu:     mu,
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
		line := data[:idx+1]
		data = data[idx+1:]
		if _, err := fmt.Fprint(pw.dest, pw.prefix+string(line)); err != nil {
			return total, err
		}
	}

	return total, nil
}

func (pw *PrefixWriter) Flush() error {
	pw.mu.Lock()
	defer pw.mu.Unlock()

	if len(pw.partial) > 0 {
		_, err := fmt.Fprintln(pw.dest, pw.prefix+string(pw.partial))
		pw.partial = nil
		return err
	}
	return nil
}

type LineWriter struct {
	index   int
	program *tea.Program
	mu      sync.Mutex
	partial []byte
}

func NewLineWriter(index int, program *tea.Program) *LineWriter {
	return &LineWriter{
		index:   index,
		program: program,
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
		line := string(data[:idx])
		data = data[idx+1:]
		lw.program.Send(tui.ItemOutputMsg{Index: lw.index, Line: line})
	}

	return total, nil
}

func (lw *LineWriter) Flush() {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	if len(lw.partial) > 0 {
		lw.program.Send(tui.ItemOutputMsg{Index: lw.index, Line: string(lw.partial)})
		lw.partial = nil
	}
}

func colorForIndex(index int) lipgloss.Color {
	return theme.ParallelColors[index%len(theme.ParallelColors)]
}
