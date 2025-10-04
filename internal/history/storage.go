package history

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

// Storage handles file I/O operations for history entries.
type Storage struct {
	filePath string
}

// NewStorage creates a new Storage instance.
func NewStorage(filePath string) *Storage {
	return &Storage{
		filePath: filePath,
	}
}

// EnsureDir creates the history directory if it doesn't exist.
func (s *Storage) EnsureDir() error {
	dir := filepath.Dir(s.filePath)
	return os.MkdirAll(dir, 0750)
}

// Append adds a new entry to the history file.
func (s *Storage) Append(entry *Entry) (err error) {
	if err := s.EnsureDir(); err != nil {
		return fmt.Errorf("failed to create history directory: %w", err)
	}

	jsonStr, err := entry.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal entry: %w", err)
	}

	file, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open history file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close history file: %w", closeErr)
		}
	}()

	if _, err = file.WriteString(jsonStr + "\n"); err != nil {
		return fmt.Errorf("failed to write entry: %w", err)
	}

	return nil
}

// ReadAll reads all entries from the history file.
func (s *Storage) ReadAll() (entries []*Entry, err error) {
	file, openErr := os.Open(s.filePath)
	if openErr != nil {
		if os.IsNotExist(openErr) {
			return []*Entry{}, nil
		}
		return nil, fmt.Errorf("failed to open history file: %w", openErr)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close history file: %w", closeErr)
		}
	}()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		entry, scanErr := FromJSON(line)
		if scanErr != nil {
			continue
		}

		entries = append(entries, entry)
	}

	if scanErr := scanner.Err(); scanErr != nil {
		return nil, fmt.Errorf("failed to read history file: %w", scanErr)
	}

	return entries, nil
}

// Clear removes all history entries.
func (s *Storage) Clear() error {
	if err := os.Remove(s.filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear history: %w", err)
	}
	return nil
}

// GetPath returns the history file path.
func (s *Storage) GetPath() string {
	return s.filePath
}
