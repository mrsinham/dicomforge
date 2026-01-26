package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/mrsinham/dicomforge/cmd/dicomforge/wizard"
	"github.com/mrsinham/dicomforge/cmd/dicomforge/wizard/components"
)

// SummaryAction represents the action selected on the summary screen
type SummaryAction int

const (
	// SummaryActionBack returns to the previous screen
	SummaryActionBack SummaryAction = iota
	// SummaryActionGenerate starts DICOM generation
	SummaryActionGenerate
	// SummaryActionSaveConfig saves configuration to YAML file
	SummaryActionSaveConfig
	// SummaryActionCancel exits the wizard
	SummaryActionCancel
)

const (
	actionBack       = "back"
	actionGenerate   = "generate"
	actionSaveConfig = "save_config"
	actionCancel     = "cancel"
)

var (
	summaryPanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2)

	summaryTitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("63")).
		Bold(true).
		MarginBottom(1)

	summaryLabelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("244"))

	summaryValueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Bold(true)

	treeStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("244"))

	treeFolderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("33"))

	treeNameStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	cliCommandStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252")).
		Padding(0, 1)
)

// SummaryScreen displays a summary of wizard configuration before generation
type SummaryScreen struct {
	form      *huh.Form
	state     *wizard.WizardState
	action    string
	done      bool
	cancelled bool
	width     int
	height    int
}

// NewSummaryScreen creates a new summary screen
func NewSummaryScreen(state *wizard.WizardState) *SummaryScreen {
	s := &SummaryScreen{
		state:  state,
		action: actionGenerate, // Default action
	}

	s.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("action").
				Title("Select an action").
				Options(
					huh.NewOption("Generate DICOM files", actionGenerate),
					huh.NewOption("Save configuration to YAML", actionSaveConfig),
					huh.NewOption("Back to edit", actionBack),
					huh.NewOption("Cancel and exit", actionCancel),
				).
				Value(&s.action),
		),
	).WithShowHelp(false)

	return s
}

// Init implements tea.Model
func (s *SummaryScreen) Init() tea.Cmd {
	return s.form.Init()
}

// Update implements tea.Model
func (s *SummaryScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			s.cancelled = true
			return s, tea.Quit
		case "esc":
			// Esc goes back instead of cancelling
			s.action = actionBack
			s.done = true
			return s, nil
		}
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
	}

	form, cmd := s.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		s.form = f
	}

	if s.form.State == huh.StateCompleted {
		s.done = true
	}

	return s, cmd
}

// View implements tea.Model
func (s *SummaryScreen) View() string {
	if s.cancelled {
		return "Cancelled.\n"
	}

	title := components.TitleStyle.Render("SUMMARY - Review Configuration")

	// Build left panel (parameter summary)
	leftPanel := s.buildParameterSummary()

	// Build right panel (tree view)
	rightPanel := s.buildTreeView()

	// Join panels side by side
	panelWidth := 45
	leftStyled := summaryPanelStyle.Width(panelWidth).Render(leftPanel)
	rightStyled := summaryPanelStyle.Width(panelWidth).Render(rightPanel)
	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftStyled, "  ", rightStyled)

	// Build CLI command section
	cliSection := s.buildCLICommand()

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		panels,
		"",
		cliSection,
		"",
		s.form.View(),
		"",
		"Enter: Select action | Esc: Back",
	)

	return content
}

// buildParameterSummary builds the left panel showing parameter summary
func (s *SummaryScreen) buildParameterSummary() string {
	var sb strings.Builder

	sb.WriteString(summaryTitleStyle.Render("Configuration Summary"))
	sb.WriteString("\n\n")

	// Calculate totals
	totalStudies := 0
	totalSeries := 0
	for _, patient := range s.state.Patients {
		totalStudies += len(patient.Studies)
		for _, study := range patient.Studies {
			totalSeries += len(study.Series)
		}
	}

	// If no patients configured yet, use global counts
	if len(s.state.Patients) == 0 {
		totalStudies = s.state.Global.NumPatients * s.state.Global.StudiesPerPatient
		totalSeries = totalStudies * s.state.Global.SeriesPerStudy
	}

	params := []struct {
		label string
		value string
	}{
		{"Modality", s.state.Global.Modality},
		{"Total Images", fmt.Sprintf("%d", s.state.Global.TotalImages)},
		{"Total Size", s.state.Global.TotalSize},
		{"Output Directory", s.state.Global.OutputDir},
		{"Number of Patients", fmt.Sprintf("%d", s.state.Global.NumPatients)},
		{"Total Studies", fmt.Sprintf("%d", totalStudies)},
		{"Total Series", fmt.Sprintf("%d", totalSeries)},
	}

	for _, p := range params {
		sb.WriteString(summaryLabelStyle.Render(p.label + ": "))
		sb.WriteString(summaryValueStyle.Render(p.value))
		sb.WriteString("\n")
	}

	return sb.String()
}

// buildTreeView builds the right panel showing the tree structure
func (s *SummaryScreen) buildTreeView() string {
	var sb strings.Builder

	sb.WriteString(summaryTitleStyle.Render("Structure Preview"))
	sb.WriteString("\n\n")

	// Folder icon
	folder := treeFolderStyle.Render("[DIR]")

	// Root output directory
	sb.WriteString(folder)
	sb.WriteString(" ")
	sb.WriteString(treeNameStyle.Render(s.state.Global.OutputDir + "/"))
	sb.WriteString("\n")

	// Build tree for patients
	patients := s.state.Patients
	if len(patients) == 0 {
		// Generate preview structure if not configured yet
		patients = s.generatePreviewPatients()
	}

	numPatients := len(patients)
	for pi, patient := range patients {
		isLastPatient := pi == numPatients-1
		patientPrefix := getTreePrefix(isLastPatient)

		// Extract short name for display
		shortName := patient.Name
		if len(shortName) > 15 {
			shortName = shortName[:15] + "..."
		}

		// Patient folder
		sb.WriteString(treeStyle.Render(patientPrefix))
		sb.WriteString(" ")
		sb.WriteString(folder)
		sb.WriteString(" ")
		sb.WriteString(treeNameStyle.Render(fmt.Sprintf("PT%06d", pi)))
		sb.WriteString(treeStyle.Render(fmt.Sprintf(" (%s)", shortName)))
		sb.WriteString("\n")

		// Studies
		studies := patient.Studies
		if len(studies) == 0 {
			studies = make([]wizard.StudyConfig, s.state.Global.StudiesPerPatient)
		}

		numStudies := len(studies)
		for si := range studies {
			isLastStudy := si == numStudies-1
			studyPrefix := getChildPrefix(isLastPatient, isLastStudy)

			// Study folder
			sb.WriteString(treeStyle.Render(studyPrefix))
			sb.WriteString(" ")
			sb.WriteString(folder)
			sb.WriteString(" ")
			sb.WriteString(treeNameStyle.Render(fmt.Sprintf("ST%06d", si)))
			sb.WriteString("\n")

			// Series
			series := studies[si].Series
			if len(series) == 0 {
				series = make([]wizard.SeriesConfig, s.state.Global.SeriesPerStudy)
			}

			numSeries := len(series)
			for sei := range series {
				isLastSeries := sei == numSeries-1
				seriesPrefix := getGrandchildPrefix(isLastPatient, isLastStudy, isLastSeries)

				// Series folder
				sb.WriteString(treeStyle.Render(seriesPrefix))
				sb.WriteString(" ")
				sb.WriteString(folder)
				sb.WriteString(" ")
				sb.WriteString(treeNameStyle.Render(fmt.Sprintf("SE%06d", sei)))
				sb.WriteString("\n")
			}
		}

		// Limit display for large hierarchies
		if pi >= 2 && numPatients > 3 {
			sb.WriteString(treeStyle.Render("    ... and "))
			sb.WriteString(summaryValueStyle.Render(fmt.Sprintf("%d", numPatients-3)))
			sb.WriteString(treeStyle.Render(" more patients"))
			sb.WriteString("\n")
			break
		}
	}

	return sb.String()
}

// getTreePrefix returns the prefix for a tree node
func getTreePrefix(isLast bool) string {
	if isLast {
		return "└──"
	}
	return "├──"
}

// getChildPrefix returns the prefix for a child node
func getChildPrefix(parentIsLast, isLast bool) string {
	var prefix string
	if parentIsLast {
		prefix = "    "
	} else {
		prefix = "│   "
	}
	if isLast {
		return prefix + "└──"
	}
	return prefix + "├──"
}

// getGrandchildPrefix returns the prefix for a grandchild node
func getGrandchildPrefix(grandparentIsLast, parentIsLast, isLast bool) string {
	var prefix string
	if grandparentIsLast {
		prefix = "    "
	} else {
		prefix = "│   "
	}
	if parentIsLast {
		prefix += "    "
	} else {
		prefix += "│   "
	}
	if isLast {
		return prefix + "└──"
	}
	return prefix + "├──"
}

// generatePreviewPatients generates preview patient structures
func (s *SummaryScreen) generatePreviewPatients() []wizard.PatientConfig {
	patients := make([]wizard.PatientConfig, s.state.Global.NumPatients)
	for i := range patients {
		patients[i] = wizard.PatientConfig{
			Name: generateDefaultPatientName(i),
			ID:   fmt.Sprintf("PAT%06d", i+1),
		}
	}
	return patients
}

// buildCLICommand builds the CLI command equivalent section
func (s *SummaryScreen) buildCLICommand() string {
	var sb strings.Builder

	sb.WriteString(summaryTitleStyle.Render("Equivalent CLI Command"))
	sb.WriteString("\n\n")

	// Build the command
	cmd := s.generateCLICommand()

	sb.WriteString(cliCommandStyle.Render(cmd))

	return sb.String()
}

// generateCLICommand generates the equivalent CLI command from wizard state
func (s *SummaryScreen) generateCLICommand() string {
	var parts []string

	parts = append(parts, "dicomforge")

	// Modality
	if s.state.Global.Modality != "" {
		parts = append(parts, fmt.Sprintf("--modality %s", s.state.Global.Modality))
	}

	// Total images
	if s.state.Global.TotalImages > 0 {
		parts = append(parts, fmt.Sprintf("--num-images %d", s.state.Global.TotalImages))
	}

	// Total size
	if s.state.Global.TotalSize != "" {
		parts = append(parts, fmt.Sprintf("--total-size %s", s.state.Global.TotalSize))
	}

	// Output directory
	if s.state.Global.OutputDir != "" && s.state.Global.OutputDir != "dicom_series" {
		parts = append(parts, fmt.Sprintf("--output %s", s.state.Global.OutputDir))
	}

	// Number of patients
	if s.state.Global.NumPatients > 1 {
		parts = append(parts, fmt.Sprintf("--num-patients %d", s.state.Global.NumPatients))
	}

	// Studies per patient
	if s.state.Global.StudiesPerPatient > 1 {
		parts = append(parts, fmt.Sprintf("--studies-per-patient %d", s.state.Global.StudiesPerPatient))
	}

	// Series per study
	if s.state.Global.SeriesPerStudy > 1 {
		parts = append(parts, fmt.Sprintf("--series-per-study %d", s.state.Global.SeriesPerStudy))
	}

	// Seed if set
	if s.state.Global.Seed != 0 {
		parts = append(parts, fmt.Sprintf("--seed %d", s.state.Global.Seed))
	}

	return strings.Join(parts, " ")
}

// Done returns true if the form was completed
func (s *SummaryScreen) Done() bool {
	return s.done
}

// Cancelled returns true if the user cancelled
func (s *SummaryScreen) Cancelled() bool {
	return s.cancelled
}

// Action returns the selected action
func (s *SummaryScreen) Action() SummaryAction {
	switch s.action {
	case actionBack:
		return SummaryActionBack
	case actionGenerate:
		return SummaryActionGenerate
	case actionSaveConfig:
		return SummaryActionSaveConfig
	case actionCancel:
		return SummaryActionCancel
	default:
		return SummaryActionGenerate
	}
}
