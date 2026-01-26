package screens

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/mrsinham/dicomforge/cmd/dicomforge/wizard/components"
	"github.com/mrsinham/dicomforge/cmd/dicomforge/wizard/types"
)

// SeriesScreen configures a single series
type SeriesScreen struct {
	form             *huh.Form
	helpPanel        *components.HelpPanel
	series           *types.SeriesConfig
	seriesIndex      int    // 0-based index
	totalSeries      int    // total number of series
	studyDescription string // study description for display
	modality         string // modality for default generation
	bodyPart         string // body part for protocol generation
	totalImages      int    // total images to distribute
	imageCountStr    string // string value for form binding
	done             bool
	cancelled        bool
	width            int
	height           int
}

// NewSeriesScreen creates a new series configuration screen
func NewSeriesScreen(series *types.SeriesConfig, index, total int, studyDescription, modality, bodyPart string, totalImages int) *SeriesScreen {
	// Calculate default image count (distribute evenly)
	defaultImageCount := totalImages / total
	if defaultImageCount < 1 {
		defaultImageCount = 1
	}

	// Set defaults if not provided
	if series.Orientation == "" {
		series.Orientation = "AXIAL"
	}
	if series.Description == "" {
		series.Description = generateDefaultSeriesDescription(modality, series.Orientation)
	}
	if series.Protocol == "" {
		series.Protocol = generateDefaultProtocol(modality, bodyPart, series.Orientation)
	}
	if series.ImageCount == 0 {
		series.ImageCount = defaultImageCount
	}

	s := &SeriesScreen{
		helpPanel:        components.NewHelpPanel(),
		series:           series,
		seriesIndex:      index,
		totalSeries:      total,
		studyDescription: studyDescription,
		modality:         modality,
		bodyPart:         bodyPart,
		totalImages:      totalImages,
		imageCountStr:    strconv.Itoa(series.ImageCount),
	}

	// Create form
	s.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("series_description").
				Title("Series Description").
				Description("Description of the series (e.g., T1 SAG, T2 AX)").
				Value(&series.Description).
				Validate(func(str string) error {
					if str == "" {
						return fmt.Errorf("series description is required")
					}
					return nil
				}),

			huh.NewInput().
				Key("protocol").
				Title("Protocol Name").
				Description("Protocol name (e.g., BRAIN_T1_SAG)").
				Value(&series.Protocol),

			huh.NewSelect[string]().
				Key("orientation").
				Title("Orientation").
				Options(
					huh.NewOption("Axial", "AXIAL"),
					huh.NewOption("Sagittal", "SAGITTAL"),
					huh.NewOption("Coronal", "CORONAL"),
				).
				Value(&series.Orientation),

			huh.NewInput().
				Key("images_in_series").
				Title("Images in Series").
				Description("Number of images in this series").
				Value(&s.imageCountStr).
				Validate(validateImageCount),
		),
	).WithShowHelp(false).WithShowErrors(true)

	return s
}

func generateDefaultSeriesDescription(modality, orientation string) string {
	// Generate description based on modality and orientation
	orientationShort := map[string]string{
		"AXIAL":    "AX",
		"SAGITTAL": "SAG",
		"CORONAL":  "COR",
	}

	modalitySequence := map[string]string{
		"MR": "T1",
		"CT": "Standard",
		"CR": "Standard",
		"DX": "Standard",
		"US": "Standard",
		"MG": "Standard",
	}

	seq := modalitySequence[modality]
	if seq == "" {
		seq = "Standard"
	}

	orient := orientationShort[orientation]
	if orient == "" {
		orient = orientation
	}

	return fmt.Sprintf("%s %s", seq, orient)
}

func generateDefaultProtocol(modality, bodyPart, orientation string) string {
	// Generate protocol from modality + body part + orientation
	bp := bodyPart
	if bp == "" {
		bp = "HEAD"
	}

	orient := orientation
	if orient == "" {
		orient = "AXIAL"
	}

	return fmt.Sprintf("%s_%s_%s", bp, modality, orient)
}

func validateImageCount(s string) error {
	if s == "" {
		return fmt.Errorf("images in series is required")
	}
	count, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("must be a valid number")
	}
	if count < 1 {
		return fmt.Errorf("must be at least 1")
	}
	return nil
}

// Init implements tea.Model
func (s *SeriesScreen) Init() tea.Cmd {
	return s.form.Init()
}

// Update implements tea.Model
func (s *SeriesScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		// Parse image count from string
		if count, err := strconv.Atoi(s.imageCountStr); err == nil {
			s.series.ImageCount = count
		}
		s.done = true
	}

	return s, cmd
}

// View implements tea.Model
func (s *SeriesScreen) View() string {
	if s.cancelled {
		return "Cancelled.\n"
	}

	title := components.TitleStyle.Render(fmt.Sprintf("SERIES %d/%d", s.seriesIndex+1, s.totalSeries))
	subtitle := components.SubtitleStyle.Render(fmt.Sprintf("Study: %s", s.studyDescription))

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
func (s *SeriesScreen) Done() bool { return s.done }

// Cancelled returns true if the user cancelled
func (s *SeriesScreen) Cancelled() bool { return s.cancelled }

// Series returns the configured series
func (s *SeriesScreen) Series() *types.SeriesConfig { return s.series }

// BulkSeriesChoice represents the user's choice for remaining series
type BulkSeriesChoice int

const (
	// BulkSeriesGenerate indicates series should be generated automatically
	BulkSeriesGenerate BulkSeriesChoice = iota
	// BulkSeriesConfigure indicates each series should be configured individually
	BulkSeriesConfigure
)

// BulkSeriesScreen shows options for remaining series after first is configured
type BulkSeriesScreen struct {
	form             *huh.Form
	choice           string
	studyDescription string
	done             bool
	cancelled        bool
	width            int
	height           int
}

// NewBulkSeriesScreen creates a new bulk series choice screen
func NewBulkSeriesScreen(remainingCount int, studyDescription string) *BulkSeriesScreen {
	s := &BulkSeriesScreen{
		choice:           choiceGenerate,
		studyDescription: studyDescription,
	}

	s.form = huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Configure Remaining Series").
				Description(fmt.Sprintf("You have configured Series 1 for study %s. For the remaining %d series:", studyDescription, remainingCount)),

			huh.NewSelect[string]().
				Key("bulk_series_choice").
				Title("What would you like to do?").
				Options(
					huh.NewOption("Generate automatically (default values)", choiceGenerate),
					huh.NewOption("Configure each one individually", choiceConfigure),
				).
				Value(&s.choice),
		),
	).WithShowHelp(false)

	return s
}

// Init implements tea.Model
func (s *BulkSeriesScreen) Init() tea.Cmd {
	return s.form.Init()
}

// Update implements tea.Model
func (s *BulkSeriesScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
func (s *BulkSeriesScreen) View() string {
	if s.cancelled {
		return "Cancelled.\n"
	}

	title := components.TitleStyle.Render("REMAINING SERIES")
	subtitle := components.SubtitleStyle.Render(fmt.Sprintf("Study: %s", s.studyDescription))

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
func (s *BulkSeriesScreen) Done() bool { return s.done }

// Cancelled returns true if the user cancelled
func (s *BulkSeriesScreen) Cancelled() bool { return s.cancelled }

// Choice returns the selected bulk choice
func (s *BulkSeriesScreen) Choice() BulkSeriesChoice {
	if s.choice == choiceConfigure {
		return BulkSeriesConfigure
	}
	return BulkSeriesGenerate
}
