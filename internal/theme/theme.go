package theme

import "github.com/charmbracelet/lipgloss"

var (
	Pink   = lipgloss.Color("212")
	Purple = lipgloss.Color("141")
	Cyan   = lipgloss.Color("43")
	White  = lipgloss.Color("255")
	Gray   = lipgloss.Color("240")
	Dim    = lipgloss.Color("238")
	Muted  = lipgloss.Color("245")

	ParallelColors = []lipgloss.Color{
		lipgloss.Color("212"), // pink
		lipgloss.Color("43"),  // cyan
		lipgloss.Color("141"), // purple
		lipgloss.Color("220"), // yellow
		lipgloss.Color("82"),  // green
		lipgloss.Color("208"), // orange
		lipgloss.Color("51"),  // bright cyan
		lipgloss.Color("199"), // magenta
	}
)
