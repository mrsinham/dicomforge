package components

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		MarginBottom(1)
)
