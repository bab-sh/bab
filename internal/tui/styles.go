package tui

import "github.com/charmbracelet/lipgloss"

var (
	cursorColor   = lipgloss.Color("198")
	selectedColor = lipgloss.Color("255")
	normalColor   = lipgloss.Color("250")
	dimColor      = lipgloss.Color("240")
	statusColor   = lipgloss.Color("240")
	groupColor    = lipgloss.Color("63")
)

var (
	// CursorStyle is the style for the selection cursor indicator.
	CursorStyle = lipgloss.NewStyle().Foreground(cursorColor).Bold(true)
	// SelectedTaskStyle is the style for the currently selected task name.
	SelectedTaskStyle = lipgloss.NewStyle().Foreground(selectedColor).Bold(true)
	// NormalTaskStyle is the style for unselected task names.
	NormalTaskStyle = lipgloss.NewStyle().Foreground(normalColor)
	// DescriptionStyle is the style for task descriptions.
	DescriptionStyle = lipgloss.NewStyle().Foreground(dimColor)
	// GroupStyle is the style for task group names.
	GroupStyle = lipgloss.NewStyle().Foreground(groupColor)
	// StatusStyle is the style for the status bar.
	StatusStyle = lipgloss.NewStyle().Foreground(statusColor)
)
