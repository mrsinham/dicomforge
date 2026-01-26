package screens

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/mrsinham/dicomforge/cmd/dicomforge/wizard/components"
	"github.com/mrsinham/dicomforge/cmd/dicomforge/wizard/types"
)

// PatientScreen configures a single patient
type PatientScreen struct {
	form           *huh.Form
	helpPanel      *components.HelpPanel
	patient        *types.PatientConfig
	patientIndex   int  // 0-based index
	totalPatients  int  // total number of patients
	acceptDefaults bool // Accept defaults for studies
	done           bool
	cancelled      bool
	width          int
	height         int
}

// NewPatientScreen creates a new patient configuration screen
func NewPatientScreen(patient *types.PatientConfig, index, total int) *PatientScreen {
	// Set defaults if not provided
	if patient.Name == "" {
		patient.Name = generateDefaultPatientName(index)
	}
	if patient.ID == "" {
		patient.ID = fmt.Sprintf("PAT%06d", index+1)
	}
	if patient.BirthDate == "" {
		patient.BirthDate = "1980-01-15" // Default date
	}
	if patient.Sex == "" {
		patient.Sex = "M"
	}

	s := &PatientScreen{
		helpPanel:     components.NewHelpPanel(),
		patient:       patient,
		patientIndex:  index,
		totalPatients: total,
	}

	// Create form
	s.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("patient_name").
				Title("Patient Name").
				Description("Format: FAMILY^Given").
				Value(&patient.Name).
				Validate(validatePatientName),

			huh.NewInput().
				Key("patient_id").
				Title("Patient ID").
				Value(&patient.ID).
				Validate(func(str string) error {
					if str == "" {
						return fmt.Errorf("patient ID is required")
					}
					return nil
				}),

			huh.NewInput().
				Key("birth_date").
				Title("Birth Date").
				Description("Format: YYYY-MM-DD").
				Value(&patient.BirthDate).
				Validate(validateDate),

			huh.NewSelect[string]().
				Key("sex").
				Title("Sex").
				Options(
					huh.NewOption("Male", "M"),
					huh.NewOption("Female", "F"),
					huh.NewOption("Other", "O"),
				).
				Value(&patient.Sex),

			huh.NewConfirm().
				Key("accept_defaults").
				Title("Accept defaults for all studies of this patient?").
				Value(&s.acceptDefaults),
		),
	).WithShowHelp(false).WithShowErrors(true)

	return s
}

func generateDefaultPatientName(index int) string {
	// Simple default names
	names := []string{
		"SMITH^John", "JOHNSON^Mary", "WILLIAMS^James",
		"BROWN^Patricia", "JONES^Robert", "GARCIA^Linda",
		"MILLER^Michael", "DAVIS^Barbara", "RODRIGUEZ^William",
		"MARTINEZ^Elizabeth",
	}
	return names[index%len(names)]
}

func validatePatientName(s string) error {
	if s == "" {
		return fmt.Errorf("patient name is required")
	}
	// Should contain ^ for DICOM format
	// But allow flexibility
	return nil
}

func validateDate(s string) error {
	if s == "" {
		return fmt.Errorf("date is required")
	}
	_, err := time.Parse("2006-01-02", s)
	if err != nil {
		return fmt.Errorf("invalid date format, use YYYY-MM-DD")
	}
	return nil
}

// Init implements tea.Model
func (s *PatientScreen) Init() tea.Cmd {
	return s.form.Init()
}

// Update implements tea.Model
func (s *PatientScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	form, cmd := s.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		s.form = f
	}

	// Update help panel
	if focused := s.form.GetFocusedField(); focused != nil {
		s.helpPanel.SetField(focused.GetKey())
	}

	if s.form.State == huh.StateCompleted {
		s.done = true
	}

	return s, cmd
}

// View implements tea.Model
func (s *PatientScreen) View() string {
	if s.cancelled {
		return "Cancelled.\n"
	}

	title := components.TitleStyle.Render(fmt.Sprintf("PATIENT %d/%d", s.patientIndex+1, s.totalPatients))

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		s.form.View(),
		"",
		s.helpPanel.View(),
		"",
		"Tab: Next field | Enter: Submit | Esc: Cancel",
	)

	return content
}

// Done returns true if the form was completed
func (s *PatientScreen) Done() bool { return s.done }

// Cancelled returns true if the user cancelled
func (s *PatientScreen) Cancelled() bool { return s.cancelled }

// Patient returns the configured patient
func (s *PatientScreen) Patient() *types.PatientConfig { return s.patient }

// AcceptDefaults returns true if the user chose to accept defaults for studies
func (s *PatientScreen) AcceptDefaults() bool { return s.acceptDefaults }

// BulkPatientChoice represents the user's choice for remaining patients
type BulkPatientChoice int

const (
	// BulkGenerate indicates patients should be generated automatically
	BulkGenerate BulkPatientChoice = iota
	// BulkConfigure indicates each patient should be configured individually
	BulkConfigure
)

// BulkPatientScreen shows options for remaining patients after first is configured
type BulkPatientScreen struct {
	form      *huh.Form
	choice    string
	done      bool
	cancelled bool
	width     int
	height    int
}

// NewBulkPatientScreen creates a new bulk patient choice screen
func NewBulkPatientScreen(remainingCount int) *BulkPatientScreen {
	s := &BulkPatientScreen{choice: "generate"}

	s.form = huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Configure Remaining Patients").
				Description(fmt.Sprintf("You have configured Patient 1. For the remaining %d patients:", remainingCount)),

			huh.NewSelect[string]().
				Key("bulk_choice").
				Title("What would you like to do?").
				Options(
					huh.NewOption("Generate automatically (random names/IDs)", "generate"),
					huh.NewOption("Configure each one individually", "configure"),
				).
				Value(&s.choice),
		),
	).WithShowHelp(false)

	return s
}

// Init implements tea.Model
func (s *BulkPatientScreen) Init() tea.Cmd {
	return s.form.Init()
}

// Update implements tea.Model
func (s *BulkPatientScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (s *BulkPatientScreen) View() string {
	if s.cancelled {
		return "Cancelled.\n"
	}

	title := components.TitleStyle.Render("REMAINING PATIENTS")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		s.form.View(),
		"",
		"Enter: Select | Esc: Cancel",
	)

	return content
}

// Done returns true if the form was completed
func (s *BulkPatientScreen) Done() bool { return s.done }

// Cancelled returns true if the user cancelled
func (s *BulkPatientScreen) Cancelled() bool { return s.cancelled }

// Choice returns the selected bulk choice
func (s *BulkPatientScreen) Choice() BulkPatientChoice {
	if s.choice == "configure" {
		return BulkConfigure
	}
	return BulkGenerate
}
