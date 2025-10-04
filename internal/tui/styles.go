package tui

import "github.com/charmbracelet/lipgloss"

var (
	cursorColor    = lipgloss.Color("198")
	dimColor       = lipgloss.Color("240")
	statusColor    = lipgloss.Color("240")
	whiteColor     = lipgloss.Color("255")
	lightGrayColor = lipgloss.Color("245")
	purpleColor    = lipgloss.Color("135")
)

var (
	// CursorStyle is the style for the selection cursor indicator.
	CursorStyle = lipgloss.NewStyle().Foreground(cursorColor).Bold(true)
	// DescriptionStyle is the style for task descriptions.
	DescriptionStyle = lipgloss.NewStyle().Foreground(dimColor)
	// ActiveSegmentStyle is the style for the current segment (white).
	ActiveSegmentStyle = lipgloss.NewStyle().Foreground(whiteColor)
	// InactiveSegmentStyle is the style for non-current segments (light gray).
	InactiveSegmentStyle = lipgloss.NewStyle().Foreground(lightGrayColor)
	// ExactMatchStyle is the style for exact substring matches (purple).
	ExactMatchStyle = lipgloss.NewStyle().Foreground(purpleColor)
	// StatusStyle is the style for the status bar.
	StatusStyle = lipgloss.NewStyle().Foreground(statusColor)
)
