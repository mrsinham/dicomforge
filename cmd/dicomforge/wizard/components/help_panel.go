package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mrsinham/dicomforge/cmd/dicomforge/wizard/help"
)

var (
	helpPanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Width(60)

	helpTitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("63")).
		Bold(true)

	helpDescStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	helpDetailStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("244"))
)

// HelpPanel displays contextual help for the current field
type HelpPanel struct {
	currentField string
	width        int
	height       int
}

// NewHelpPanel creates a new help panel
func NewHelpPanel() *HelpPanel {
	return &HelpPanel{
		width:  60,
		height: 10,
	}
}

// SetField updates which field's help to display
func (h *HelpPanel) SetField(field string) {
	h.currentField = field
}

// SetSize updates panel dimensions
func (h *HelpPanel) SetSize(width, height int) {
	h.width = width
	h.height = height
}

// View renders the help panel
func (h *HelpPanel) View() string {
	style := helpPanelStyle.Width(h.width - 4) // Compute locally, don't mutate global

	text, ok := help.Texts[h.currentField]
	if !ok {
		return style.Render("Select a field to see help")
	}

	var sb strings.Builder
	sb.WriteString("ℹ️  ")
	sb.WriteString(helpTitleStyle.Render(text.Title))
	sb.WriteString("\n\n")
	sb.WriteString(helpDescStyle.Render(text.Description))
	sb.WriteString("\n\n")
	sb.WriteString(helpDetailStyle.Render(text.Details))

	return style.Render(sb.String())
}
