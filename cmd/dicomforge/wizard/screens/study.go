package screens

import (
	"fmt"
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/mrsinham/dicomforge/cmd/dicomforge/wizard"
	"github.com/mrsinham/dicomforge/cmd/dicomforge/wizard/components"
)

// StudyScreen configures a single study
type StudyScreen struct {
	form           *huh.Form
	helpPanel      *components.HelpPanel
	study          *wizard.StudyConfig
	studyIndex     int    // 0-based index
	totalStudies   int    // total number of studies
	patientName    string // patient name for display
	modality       string // modality for default description
	acceptDefaults bool   // Accept defaults for all series
	done           bool
	cancelled      bool
	width          int
	height         int
}

// NewStudyScreen creates a new study configuration screen
func NewStudyScreen(study *wizard.StudyConfig, index, total int, patientName, modality string) *StudyScreen {
	// Set defaults if not provided
	if study.Description == "" {
		study.Description = generateDefaultStudyDescription(modality, study.BodyPart)
	}
	if study.Date == "" {
		study.Date = time.Now().Format("2006-01-02")
	}
	if study.AccessionNumber == "" {
		study.AccessionNumber = generateAccessionNumber()
	}
	if study.BodyPart == "" {
		study.BodyPart = "HEAD"
	}
	if study.Priority == "" {
		study.Priority = "ROUTINE"
	}

	s := &StudyScreen{
		helpPanel:    components.NewHelpPanel(),
		study:        study,
		studyIndex:   index,
		totalStudies: total,
		patientName:  patientName,
		modality:     modality,
	}

	// Create form
	s.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("study_description").
				Title("Study Description").
				Description("Human-readable description").
				Value(&study.Description).
				Validate(func(str string) error {
					if str == "" {
						return fmt.Errorf("study description is required")
					}
					return nil
				}),

			huh.NewInput().
				Key("study_date").
				Title("Study Date").
				Description("Format: YYYY-MM-DD").
				Value(&study.Date).
				Validate(validateStudyDate),

			huh.NewInput().
				Key("accession").
				Title("Accession Number").
				Value(&study.AccessionNumber),

			huh.NewInput().
				Key("institution").
				Title("Institution").
				Description("Hospital or imaging center").
				Value(&study.Institution),

			huh.NewInput().
				Key("department").
				Title("Department").
				Description("e.g., Radiology, Emergency").
				Value(&study.Department),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("body_part").
				Title("Body Part").
				Options(
					huh.NewOption("Head", "HEAD"),
					huh.NewOption("Brain", "BRAIN"),
					huh.NewOption("Neck", "NECK"),
					huh.NewOption("Chest", "CHEST"),
					huh.NewOption("Abdomen", "ABDOMEN"),
					huh.NewOption("Pelvis", "PELVIS"),
					huh.NewOption("Spine", "SPINE"),
					huh.NewOption("Cervical Spine", "CSPINE"),
					huh.NewOption("Thoracic Spine", "TSPINE"),
					huh.NewOption("Lumbar Spine", "LSPINE"),
					huh.NewOption("Shoulder", "SHOULDER"),
					huh.NewOption("Elbow", "ELBOW"),
					huh.NewOption("Hand", "HAND"),
					huh.NewOption("Hip", "HIP"),
					huh.NewOption("Knee", "KNEE"),
					huh.NewOption("Ankle", "ANKLE"),
					huh.NewOption("Foot", "FOOT"),
				).
				Value(&study.BodyPart),

			huh.NewSelect[string]().
				Key("priority").
				Title("Priority").
				Options(
					huh.NewOption("High (Urgent)", "HIGH"),
					huh.NewOption("Routine", "ROUTINE"),
					huh.NewOption("Low (Non-urgent)", "LOW"),
				).
				Value(&study.Priority),

			huh.NewInput().
				Key("referring_physician").
				Title("Referring Physician").
				Description("Format: DR FAMILY^Given").
				Value(&study.ReferringPhysician),

			huh.NewConfirm().
				Key("accept_defaults").
				Title("Accept defaults for all series of this study?").
				Value(&s.acceptDefaults),
		),
	).WithShowHelp(false).WithShowErrors(true)

	return s
}

func generateDefaultStudyDescription(modality, bodyPart string) string {
	// Generate description based on modality and body part
	bp := bodyPart
	if bp == "" {
		bp = "HEAD"
	}

	modalityNames := map[string]string{
		"MR": "MRI",
		"CT": "CT Scan",
		"CR": "X-Ray",
		"DX": "Digital X-Ray",
		"US": "Ultrasound",
		"MG": "Mammography",
	}

	bodyPartNames := map[string]string{
		"HEAD":     "Head",
		"BRAIN":    "Brain",
		"NECK":     "Neck",
		"CHEST":    "Chest",
		"ABDOMEN":  "Abdomen",
		"PELVIS":   "Pelvis",
		"SPINE":    "Spine",
		"CSPINE":   "Cervical Spine",
		"TSPINE":   "Thoracic Spine",
		"LSPINE":   "Lumbar Spine",
		"SHOULDER": "Shoulder",
		"ELBOW":    "Elbow",
		"HAND":     "Hand",
		"HIP":      "Hip",
		"KNEE":     "Knee",
		"ANKLE":    "Ankle",
		"FOOT":     "Foot",
	}

	modName := modalityNames[modality]
	if modName == "" {
		modName = modality
	}

	bpName := bodyPartNames[bp]
	if bpName == "" {
		bpName = bp
	}

	return fmt.Sprintf("%s %s", bpName, modName)
}

func generateAccessionNumber() string {
	return fmt.Sprintf("ACC-%06d", rand.Intn(1000000))
}

func validateStudyDate(s string) error {
	if s == "" {
		return fmt.Errorf("study date is required")
	}
	_, err := time.Parse("2006-01-02", s)
	if err != nil {
		return fmt.Errorf("invalid date format, use YYYY-MM-DD")
	}
	return nil
}

// Init implements tea.Model
func (s *StudyScreen) Init() tea.Cmd {
	return s.form.Init()
}

// Update implements tea.Model
func (s *StudyScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (s *StudyScreen) View() string {
	if s.cancelled {
		return "Cancelled.\n"
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		MarginBottom(1)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		MarginBottom(1)

	title := titleStyle.Render(fmt.Sprintf("STUDY %d/%d", s.studyIndex+1, s.totalStudies))
	subtitle := subtitleStyle.Render(fmt.Sprintf("Patient: %s", s.patientName))

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		subtitle,
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
func (s *StudyScreen) Done() bool { return s.done }

// Cancelled returns true if the user cancelled
func (s *StudyScreen) Cancelled() bool { return s.cancelled }

// Study returns the configured study
func (s *StudyScreen) Study() *wizard.StudyConfig { return s.study }

// AcceptDefaults returns true if the user chose to accept defaults for series
func (s *StudyScreen) AcceptDefaults() bool { return s.acceptDefaults }

// BulkStudyChoice represents the user's choice for remaining studies
type BulkStudyChoice int

const (
	// BulkStudyGenerate indicates studies should be generated automatically
	BulkStudyGenerate BulkStudyChoice = iota
	// BulkStudyConfigure indicates each study should be configured individually
	BulkStudyConfigure
)

// BulkStudyScreen shows options for remaining studies after first is configured
type BulkStudyScreen struct {
	form        *huh.Form
	choice      string
	patientName string
	done        bool
	cancelled   bool
	width       int
	height      int
}

// NewBulkStudyScreen creates a new bulk study choice screen
func NewBulkStudyScreen(remainingCount int, patientName string) *BulkStudyScreen {
	s := &BulkStudyScreen{
		choice:      "generate",
		patientName: patientName,
	}

	s.form = huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Configure Remaining Studies").
				Description(fmt.Sprintf("You have configured Study 1 for patient %s. For the remaining %d studies:", patientName, remainingCount)),

			huh.NewSelect[string]().
				Key("bulk_study_choice").
				Title("What would you like to do?").
				Options(
					huh.NewOption("Generate automatically (default values)", "generate"),
					huh.NewOption("Configure each one individually", "configure"),
				).
				Value(&s.choice),
		),
	).WithShowHelp(false)

	return s
}

// Init implements tea.Model
func (s *BulkStudyScreen) Init() tea.Cmd {
	return s.form.Init()
}

// Update implements tea.Model
func (s *BulkStudyScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (s *BulkStudyScreen) View() string {
	if s.cancelled {
		return "Cancelled.\n"
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("63")).
		MarginBottom(1)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		MarginBottom(1)

	title := titleStyle.Render("REMAINING STUDIES")
	subtitle := subtitleStyle.Render(fmt.Sprintf("Patient: %s", s.patientName))

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		subtitle,
		"",
		s.form.View(),
		"",
		"Enter: Select | Esc: Cancel",
	)

	return content
}

// Done returns true if the form was completed
func (s *BulkStudyScreen) Done() bool { return s.done }

// Cancelled returns true if the user cancelled
func (s *BulkStudyScreen) Cancelled() bool { return s.cancelled }

// Choice returns the selected bulk choice
func (s *BulkStudyScreen) Choice() BulkStudyChoice {
	if s.choice == "configure" {
		return BulkStudyConfigure
	}
	return BulkStudyGenerate
}
