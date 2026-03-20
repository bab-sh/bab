package picker

import (
	"charm.land/lipgloss/v2"
	"github.com/bab-sh/bab/internal/theme"
)

var (
	promptStyle           = lipgloss.NewStyle().Foreground(theme.Pink)
	inputStyle            = lipgloss.NewStyle().Foreground(theme.White)
	countStyle            = lipgloss.NewStyle().Foreground(theme.Gray)
	separatorStyle        = lipgloss.NewStyle().Foreground(theme.Dim)
	selectedIndicator     = lipgloss.NewStyle().Foreground(theme.Purple).Bold(true)
	taskNameStyle         = lipgloss.NewStyle().Foreground(theme.White)
	taskNameSelectedStyle = lipgloss.NewStyle().Foreground(theme.Purple)
	matchStyle            = lipgloss.NewStyle().Foreground(theme.Pink).Bold(true)
	descStyle             = lipgloss.NewStyle().Foreground(theme.Muted).Italic(true)
	noResultsStyle        = lipgloss.NewStyle().Foreground(theme.Gray).Italic(true)
	aliasStyle            = lipgloss.NewStyle().Foreground(theme.Cyan)
	aliasMatchStyle       = lipgloss.NewStyle().Foreground(theme.Cyan).Bold(true)
)
