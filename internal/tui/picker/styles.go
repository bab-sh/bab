package picker

import "github.com/charmbracelet/lipgloss"

const (
	colorMatch    = lipgloss.Color("212")
	colorSelected = lipgloss.Color("141")
	colorWhite    = lipgloss.Color("255")
	colorGray     = lipgloss.Color("240")
	colorDim      = lipgloss.Color("238")
	colorMuted    = lipgloss.Color("245")
)

var (
	promptStyle           = lipgloss.NewStyle().Foreground(colorMatch)
	inputStyle            = lipgloss.NewStyle().Foreground(colorWhite)
	countStyle            = lipgloss.NewStyle().Foreground(colorGray)
	separatorStyle        = lipgloss.NewStyle().Foreground(colorDim)
	selectedIndicator     = lipgloss.NewStyle().Foreground(colorSelected).Bold(true)
	taskNameStyle         = lipgloss.NewStyle().Foreground(colorWhite)
	taskNameSelectedStyle = lipgloss.NewStyle().Foreground(colorSelected)
	matchStyle            = lipgloss.NewStyle().Foreground(colorMatch).Bold(true)
	descStyle             = lipgloss.NewStyle().Foreground(colorMuted).Italic(true)
	noResultsStyle        = lipgloss.NewStyle().Foreground(colorGray).Italic(true)
)
