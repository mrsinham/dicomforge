package screens

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mrsinham/dicomforge/cmd/dicomforge/wizard/components"
)

// ProgressMsg is sent to update the progress screen during generation
type ProgressMsg struct {
	Current int    // Current file number
	Total   int    // Total files to generate
	Path    string // Current file path being written
}

// CompletionMsg is sent when generation completes successfully
type CompletionMsg struct {
	TotalFiles int           // Total number of files created
	TotalSize  int64         // Total size in bytes
	Duration   time.Duration // Time taken
	OutputDir  string        // Output directory path
}

// ErrorMsg is sent when an error occurs during generation
type ErrorMsg struct {
	Error error
}

var (
	progressBarStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("63"))

	progressBarEmptyStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))

	progressPercentStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("63")).
				Bold(true)

	progressFileStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("244"))

	progressElapsedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("244"))

	cancelHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
)

// ProgressScreen displays generation progress
type ProgressScreen struct {
	current   int
	total     int
	path      string
	startTime time.Time
	cancelled bool
	width     int
	height    int
}

// NewProgressScreen creates a new progress screen
func NewProgressScreen(total int) *ProgressScreen {
	return &ProgressScreen{
		current:   0,
		total:     total,
		startTime: time.Now(),
	}
}

// Init implements tea.Model
func (s *ProgressScreen) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (s *ProgressScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			s.cancelled = true
			return s, tea.Quit
		}
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
	case ProgressMsg:
		s.current = msg.Current
		s.total = msg.Total
		s.path = msg.Path
	}

	return s, nil
}

// View implements tea.Model
func (s *ProgressScreen) View() string {
	if s.cancelled {
		return "Cancelled.\n"
	}

	title := components.TitleStyle.Render("Generating DICOM files...")

	// Calculate progress
	var percent float64
	if s.total > 0 {
		percent = float64(s.current) / float64(s.total) * 100
	}

	// Build progress bar
	barWidth := 40
	if s.width > 60 {
		barWidth = s.width / 2
		if barWidth > 60 {
			barWidth = 60
		}
	}
	progressBar := s.renderProgressBar(percent, barWidth)

	// Percentage display
	percentStr := progressPercentStyle.Render(fmt.Sprintf("%d%%", int(percent)))

	// File counter
	fileCounter := progressFileStyle.Render(fmt.Sprintf("File %d/%d", s.current, s.total))

	// Current path
	var pathDisplay string
	if s.path != "" {
		// Truncate path if too long
		displayPath := s.path
		maxPathLen := barWidth
		if len(displayPath) > maxPathLen {
			displayPath = "..." + displayPath[len(displayPath)-maxPathLen+3:]
		}
		pathDisplay = progressFileStyle.Render(displayPath)
	}

	// Elapsed time
	elapsed := time.Since(s.startTime)
	elapsedStr := progressElapsedStyle.Render(fmt.Sprintf("Elapsed: %.1fs", elapsed.Seconds()))

	// Cancel hint
	cancelHint := cancelHintStyle.Render("Press Ctrl+C to cancel")

	// Build the view
	var sb strings.Builder
	sb.WriteString(title)
	sb.WriteString("\n\n")
	sb.WriteString(progressBar)
	sb.WriteString(" ")
	sb.WriteString(percentStr)
	sb.WriteString("\n\n")
	sb.WriteString(fileCounter)
	if pathDisplay != "" {
		sb.WriteString(": ")
		sb.WriteString(pathDisplay)
	}
	sb.WriteString("\n")
	sb.WriteString(elapsedStr)
	sb.WriteString("\n\n")
	sb.WriteString(cancelHint)

	return sb.String()
}

// renderProgressBar creates a visual progress bar
func (s *ProgressScreen) renderProgressBar(percent float64, width int) string {
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled

	bar := progressBarStyle.Render("[" + strings.Repeat("█", filled))
	bar += progressBarEmptyStyle.Render(strings.Repeat("░", empty) + "]")

	return bar
}

// Cancelled returns true if the user cancelled
func (s *ProgressScreen) Cancelled() bool {
	return s.cancelled
}

// SetProgress updates the progress (for external updates)
func (s *ProgressScreen) SetProgress(current, total int, path string) {
	s.current = current
	s.total = total
	s.path = path
}

// Completion screen styles
var (
	completionSuccessStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("42")).
				Bold(true)

	completionLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("244"))

	completionValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Bold(true)

	completionHintStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("244")).
				Italic(true)

	completionCommandStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("236")).
				Foreground(lipgloss.Color("252")).
				Padding(0, 1)

	completionButtonFocusedStyle = lipgloss.NewStyle().
					Background(lipgloss.Color("33")).
					Foreground(lipgloss.Color("255")).
					Padding(0, 2).
					Bold(true)
)

// CompletionScreen displays the completion summary
type CompletionScreen struct {
	totalFiles int
	totalSize  int64
	duration   time.Duration
	outputDir  string
	done       bool
	width      int
	height     int
}

// NewCompletionScreen creates a new completion screen
func NewCompletionScreen(msg CompletionMsg) *CompletionScreen {
	return &CompletionScreen{
		totalFiles: msg.TotalFiles,
		totalSize:  msg.TotalSize,
		duration:   msg.Duration,
		outputDir:  msg.OutputDir,
	}
}

// Init implements tea.Model
func (s *CompletionScreen) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (s *CompletionScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "enter", "q":
			s.done = true
			return s, tea.Quit
		}
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
	}

	return s, nil
}

// View implements tea.Model
func (s *CompletionScreen) View() string {
	var sb strings.Builder

	// Success header
	successIcon := completionSuccessStyle.Render("✓")
	successText := completionSuccessStyle.Render("Generation complete!")
	sb.WriteString(successIcon)
	sb.WriteString(" ")
	sb.WriteString(successText)
	sb.WriteString("\n\n")

	// Summary section
	sb.WriteString(components.TitleStyle.Render("Summary:"))
	sb.WriteString("\n")

	// Stats
	stats := []struct {
		label string
		value string
	}{
		{"Files created", fmt.Sprintf("%d", s.totalFiles)},
		{"Total size", formatSize(s.totalSize)},
		{"Duration", fmt.Sprintf("%.1fs", s.duration.Seconds())},
		{"Output", s.outputDir},
	}

	for _, stat := range stats {
		sb.WriteString("  ")
		sb.WriteString(completionLabelStyle.Render(stat.label + ":"))
		sb.WriteString(" ")
		sb.WriteString(completionValueStyle.Render(stat.value))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	// Next steps
	sb.WriteString(components.TitleStyle.Render("Next steps:"))
	sb.WriteString("\n")

	listCmd := completionCommandStyle.Render(fmt.Sprintf("ls -la %s", s.outputDir))
	sb.WriteString("  • View files: ")
	sb.WriteString(listCmd)
	sb.WriteString("\n")

	// Find first DICOM file path pattern
	dcmPath := fmt.Sprintf("%s/PT000000/ST000000/SE000000/IM000000", s.outputDir)
	validateCmd := completionCommandStyle.Render(fmt.Sprintf("dcmdump %s", dcmPath))
	sb.WriteString("  • Validate:   ")
	sb.WriteString(validateCmd)
	sb.WriteString("\n\n")

	// Exit button
	exitButton := completionButtonFocusedStyle.Render("Exit")
	sb.WriteString(exitButton)
	sb.WriteString("\n\n")
	sb.WriteString(completionHintStyle.Render("Press Enter or q to exit"))

	return sb.String()
}

// Done returns true if the user is finished
func (s *CompletionScreen) Done() bool {
	return s.done
}

// formatSize formats bytes as human-readable size
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// ErrorScreen displays an error that occurred during generation
type ErrorScreen struct {
	err    error
	done   bool
	width  int
	height int
}

var (
	errorTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	errorMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	errorHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Italic(true)
)

// NewErrorScreen creates a new error screen
func NewErrorScreen(err error) *ErrorScreen {
	return &ErrorScreen{
		err: err,
	}
}

// Init implements tea.Model
func (s *ErrorScreen) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (s *ErrorScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "enter", "q":
			s.done = true
			return s, tea.Quit
		}
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
	}

	return s, nil
}

// View implements tea.Model
func (s *ErrorScreen) View() string {
	var sb strings.Builder

	// Error header
	errorIcon := errorTitleStyle.Render("✗")
	errorText := errorTitleStyle.Render("Generation failed")
	sb.WriteString(errorIcon)
	sb.WriteString(" ")
	sb.WriteString(errorText)
	sb.WriteString("\n\n")

	// Error message
	sb.WriteString(components.TitleStyle.Render("Error:"))
	sb.WriteString("\n")
	sb.WriteString("  ")
	sb.WriteString(errorMessageStyle.Render(s.err.Error()))
	sb.WriteString("\n\n")

	// Exit hint
	sb.WriteString(errorHintStyle.Render("Press Enter or q to exit"))

	return sb.String()
}

// Done returns true if the user is finished
func (s *ErrorScreen) Done() bool {
	return s.done
}

// Error returns the error
func (s *ErrorScreen) Error() error {
	return s.err
}
