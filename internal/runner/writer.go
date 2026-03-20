package runner

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"strings"
	"sync"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bab-sh/bab/internal/theme"
	"github.com/bab-sh/bab/internal/tui"
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

type lineBuffer struct {
	partial []byte
	strip   bool
}

func (lb *lineBuffer) process(data []byte, emit func(string) error) (int, error) {
	total := len(data)
	lb.partial = append(lb.partial, data...)
	buf := lb.partial
	lb.partial = nil

	for len(buf) > 0 {
		idx := bytes.IndexByte(buf, '\n')
		if idx < 0 {
			lb.partial = append(lb.partial, buf...)
			break
		}
		line := cleanLine(string(buf[:idx]), lb.strip)
		buf = buf[idx+1:]
		if err := emit(line); err != nil {
			return total, err
		}
	}

	return total, nil
}

func (lb *lineBuffer) flush(emit func(string) error) error {
	if len(lb.partial) > 0 {
		line := cleanLine(string(lb.partial), lb.strip)
		lb.partial = nil
		return emit(line)
	}
	return nil
}

type PrefixWriter struct {
	prefix string
	dest   io.Writer
	mu     *sync.Mutex
	buf    lineBuffer
}

func NewPrefixWriter(label string, padWidth int, c color.Color, dest io.Writer, mu *sync.Mutex, strip bool) *PrefixWriter {
	style := lipgloss.NewStyle().Foreground(c)
	paddedLabel := fmt.Sprintf("%-*s", padWidth, label)
	prefix := style.Render("["+paddedLabel+"]") + " "
	return &PrefixWriter{
		prefix: prefix,
		dest:   dest,
		mu:     mu,
		buf:    lineBuffer{strip: strip},
	}
}

func (pw *PrefixWriter) Write(p []byte) (int, error) {
	pw.mu.Lock()
	defer pw.mu.Unlock()
	return pw.buf.process(p, func(line string) error {
		_, err := fmt.Fprintln(pw.dest, pw.prefix+line)
		return err
	})
}

func (pw *PrefixWriter) Flush() error {
	pw.mu.Lock()
	defer pw.mu.Unlock()
	return pw.buf.flush(func(line string) error {
		_, err := fmt.Fprintln(pw.dest, pw.prefix+line)
		return err
	})
}

type KeyLineWriter struct {
	key     string
	program *tea.Program
	mu      sync.Mutex
	buf     lineBuffer
}

func NewKeyLineWriter(key string, program *tea.Program, strip bool) *KeyLineWriter {
	return &KeyLineWriter{
		key:     key,
		program: program,
		buf:     lineBuffer{strip: strip},
	}
}

func (kw *KeyLineWriter) Write(p []byte) (int, error) {
	kw.mu.Lock()
	defer kw.mu.Unlock()
	return kw.buf.process(p, func(line string) error {
		kw.program.Send(tui.ItemOutputMsg{Key: kw.key, Line: line})
		return nil
	})
}

func (kw *KeyLineWriter) Flush() {
	kw.mu.Lock()
	defer kw.mu.Unlock()
	_ = kw.buf.flush(func(line string) error {
		kw.program.Send(tui.ItemOutputMsg{Key: kw.key, Line: line})
		return nil
	})
}

func colorForPath(path []int) color.Color {
	if len(path) == 0 {
		return theme.ParallelBaseColors[0]
	}
	bases := theme.ParallelBaseColors
	base := bases[path[0]%len(bases)]
	depth := len(path) - 1
	if depth == 0 {
		return base
	}
	return dimColor(base, depth)
}

func dimColor(c color.Color, steps int) color.Color {
	idx, ok := c.(ansi.IndexedColor)
	if !ok {
		return c
	}
	code := int(idx)
	if code < 16 || code > 231 {
		return c
	}

	ci := code - 16
	r := ci / 36
	g := (ci % 36) / 6
	b := ci % 6

	for range steps {
		r = (r * 3) / 5
		g = (g * 3) / 5
		b = (b * 3) / 5
	}

	dimmed := 16 + 36*r + 6*g + b
	return lipgloss.ANSIColor(uint8(dimmed)) //nolint:gosec // dimmed is always in [16,231]
}
