package history

import (
	"fmt"
	"path/filepath"
)

const (
	historyDirName  = ".bab"
	historyFileName = "history.jsonl"
)

// Manager manages task execution history.
type Manager struct {
	storage *Storage
}

// NewManager creates a new history manager for a specific project.
// The projectRoot should be the directory containing the Babfile.
func NewManager(projectRoot string) (*Manager, error) {
	if projectRoot == "" {
		return nil, fmt.Errorf("project root path cannot be empty")
	}

	historyPath := getHistoryPath(projectRoot)
	storage := NewStorage(historyPath)

	return &Manager{
		storage: storage,
	}, nil
}

// Record adds a new entry to the history.
func (m *Manager) Record(entry *Entry) error {
	return m.storage.Append(entry)
}

// List returns all history entries, optionally limited to the most recent N entries.
// If limit is 0 or negative, all entries are returned.
func (m *Manager) List(limit int) ([]*Entry, error) {
	entries, err := m.storage.ReadAll()
	if err != nil {
		return nil, err
	}

	if limit > 0 && len(entries) > limit {
		return entries[len(entries)-limit:], nil
	}

	return entries, nil
}

// Clear removes all history entries.
func (m *Manager) Clear() error {
	return m.storage.Clear()
}

// GetPath returns the path to the history file.
func (m *Manager) GetPath() string {
	return m.storage.GetPath()
}

// getHistoryPath returns the project-local path for the history file.
// History is stored in .bab/history.jsonl relative to the project root.
func getHistoryPath(projectRoot string) string {
	return filepath.Join(projectRoot, historyDirName, historyFileName)
}
