package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

var (
	Pink   = lipgloss.Color("212")
	Purple = lipgloss.Color("141")
	Cyan   = lipgloss.Color("43")
	White  = lipgloss.Color("255")
	Gray   = lipgloss.Color("240")
	Dim    = lipgloss.Color("238")
	Muted  = lipgloss.Color("245")

	ParallelBaseColors = []color.Color{
		lipgloss.Color("212"),
		lipgloss.Color("43"),
		lipgloss.Color("141"),
		lipgloss.Color("220"),
		lipgloss.Color("82"),
		lipgloss.Color("208"),
		lipgloss.Color("51"),
		lipgloss.Color("199"),
	}
)
