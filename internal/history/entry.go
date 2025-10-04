// Package history provides functionality for recording and managing task execution history.
package history

import (
	"encoding/json"
	"time"
)

// Status represents the execution status of a task.
type Status string

const (
	// StatusSuccess indicates the task completed successfully.
	StatusSuccess Status = "success"
	// StatusFailure indicates the task failed.
	StatusFailure Status = "failed"
)

// Entry represents a single task execution in the history.
type Entry struct {
	Task        string    `json:"task"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	Status      Status    `json:"status"`
	DurationMs  int64     `json:"duration_ms"`
	WorkDir     string    `json:"workdir"`
	Error       string    `json:"error"`
}

// NewEntry creates a new history entry.
func NewEntry(task, description, workDir string, status Status, duration time.Duration, err error) *Entry {
	entry := &Entry{
		Task:        task,
		Description: description,
		Timestamp:   time.Now(),
		Status:      status,
		DurationMs:  duration.Milliseconds(),
		WorkDir:     workDir,
		Error:       "",
	}

	if err != nil {
		entry.Error = err.Error()
	}

	return entry
}

// ToJSON converts the entry to a JSON string.
func (e *Entry) ToJSON() (string, error) {
	bytes, err := json.Marshal(e)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// FromJSON creates an entry from a JSON string.
func FromJSON(jsonStr string) (*Entry, error) {
	var entry Entry
	if err := json.Unmarshal([]byte(jsonStr), &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}
