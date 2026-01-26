package screens

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/mrsinham/dicomforge/cmd/dicomforge/wizard/components"
	"github.com/mrsinham/dicomforge/cmd/dicomforge/wizard/types"
	"github.com/mrsinham/dicomforge/internal/util"
)

// GlobalScreen is the first wizard screen for global configuration
type GlobalScreen struct {
	form      *huh.Form
	helpPanel *components.HelpPanel
	config    *types.GlobalConfig
	width     int
	height    int
	done      bool
	cancelled bool

	// String versions for form binding (huh binds to strings)
	totalImagesStr       string
	numPatientsStr       string
	studiesPerPatientStr string
	seriesPerStudyStr    string
}

// NewGlobalScreen creates a new global configuration screen
func NewGlobalScreen(config *types.GlobalConfig) *GlobalScreen {
	// Set defaults if not provided
	if config.Modality == "" {
		config.Modality = "MR"
	}
	if config.TotalImages == 0 {
		config.TotalImages = 50
	}
	if config.TotalSize == "" {
		config.TotalSize = "500MB"
	}
	if config.OutputDir == "" {
		config.OutputDir = "dicom_series"
	}
	if config.NumPatients == 0 {
		config.NumPatients = 1
	}
	if config.StudiesPerPatient == 0 {
		config.StudiesPerPatient = 1
	}
	if config.SeriesPerStudy == 0 {
		config.SeriesPerStudy = 1
	}

	s := &GlobalScreen{
		helpPanel:            components.NewHelpPanel(),
		config:               config,
		totalImagesStr:       strconv.Itoa(config.TotalImages),
		numPatientsStr:       strconv.Itoa(config.NumPatients),
		studiesPerPatientStr: strconv.Itoa(config.StudiesPerPatient),
		seriesPerStudyStr:    strconv.Itoa(config.SeriesPerStudy),
	}

	// Create form fields
	s.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("modality").
				Title("Modality").
				Options(
					huh.NewOption("MR - Magnetic Resonance", "MR"),
					huh.NewOption("CT - Computed Tomography", "CT"),
					huh.NewOption("CR - Computed Radiography", "CR"),
					huh.NewOption("DX - Digital X-Ray", "DX"),
					huh.NewOption("US - Ultrasound", "US"),
					huh.NewOption("MG - Mammography", "MG"),
				).
				Value(&config.Modality),

			huh.NewInput().
				Key("total_images").
				Title("Total Images").
				Value(&s.totalImagesStr).
				Validate(validatePositiveInt),

			huh.NewInput().
				Key("total_size").
				Title("Total Size").
				Placeholder("e.g., 500MB, 1GB").
				Value(&config.TotalSize).
				Validate(validateSize),

			huh.NewInput().
				Key("output").
				Title("Output Directory").
				Value(&config.OutputDir).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("output directory is required")
					}
					return nil
				}),
		),
		huh.NewGroup(
			huh.NewInput().
				Key("num_patients").
				Title("Number of Patients").
				Value(&s.numPatientsStr).
				Validate(validatePositiveInt),

			huh.NewInput().
				Key("studies_per_patient").
				Title("Studies per Patient").
				Value(&s.studiesPerPatientStr).
				Validate(validatePositiveInt),

			huh.NewInput().
				Key("series_per_study").
				Title("Series per Study").
				Value(&s.seriesPerStudyStr).
				Validate(validatePositiveInt),
		),
	).WithShowHelp(false).WithShowErrors(true)

	return s
}

func validatePositiveInt(s string) error {
	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("must be a number")
	}
	if n <= 0 {
		return fmt.Errorf("must be greater than 0")
	}
	return nil
}

func validateSize(s string) error {
	if s == "" {
		return fmt.Errorf("size is required")
	}
	_, err := util.ParseSize(s)
	return err
}

// Init implements tea.Model
func (s *GlobalScreen) Init() tea.Cmd {
	return s.form.Init()
}

// Update implements tea.Model
func (s *GlobalScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			s.cancelled = true
			return s, tea.Quit
		}
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.helpPanel.SetSize(msg.Width/3, msg.Height/2)
	}

	// Update form
	form, cmd := s.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		s.form = f
	}

	// Update help panel based on focused field
	focused := s.form.GetFocusedField()
	if focused != nil {
		s.helpPanel.SetField(focused.GetKey())
	}

	// Check if form is complete
	if s.form.State == huh.StateCompleted {
		s.done = true
		s.syncConfigFromForm()
	}

	return s, cmd
}

// syncConfigFromForm parses form values back to config
func (s *GlobalScreen) syncConfigFromForm() {
	// Parse string values back to ints
	if n, err := strconv.Atoi(s.totalImagesStr); err == nil {
		s.config.TotalImages = n
	}
	if n, err := strconv.Atoi(s.numPatientsStr); err == nil {
		s.config.NumPatients = n
	}
	if n, err := strconv.Atoi(s.studiesPerPatientStr); err == nil {
		s.config.StudiesPerPatient = n
	}
	if n, err := strconv.Atoi(s.seriesPerStudyStr); err == nil {
		s.config.SeriesPerStudy = n
	}
}

// View implements tea.Model
func (s *GlobalScreen) View() string {
	if s.cancelled {
		return "Cancelled.\n"
	}

	title := components.TitleStyle.Render("DICOMFORGE WIZARD - Global Configuration")

	// Layout: form on left, help panel on right
	formView := s.form.View()
	helpView := s.helpPanel.View()

	// Simple vertical layout for now
	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		formView,
		"",
		helpView,
		"",
		"Tab: Next field | Enter: Submit | Esc: Cancel",
	)

	return content
}

// Done returns true if the form was completed
func (s *GlobalScreen) Done() bool {
	return s.done
}

// Cancelled returns true if the user cancelled
func (s *GlobalScreen) Cancelled() bool {
	return s.cancelled
}

// Config returns the configured global settings
func (s *GlobalScreen) Config() *types.GlobalConfig {
	return s.config
}
