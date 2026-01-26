package wizard

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/mrsinham/dicomforge/cmd/dicomforge/wizard/components"
	"github.com/mrsinham/dicomforge/cmd/dicomforge/wizard/screens"
	"github.com/mrsinham/dicomforge/internal/dicom"
	"github.com/mrsinham/dicomforge/internal/dicom/modalities"
	"github.com/mrsinham/dicomforge/internal/util"
)

// Phase represents the current phase/screen of the wizard.
type Phase int

const (
	PhaseGlobal Phase = iota
	PhasePatient
	PhaseBulkPatient // For remaining patients
	PhaseStudy
	PhaseBulkStudy // For remaining studies
	PhaseSeries
	PhaseBulkSeries // For remaining series
	PhaseSummary
	PhaseProgress
	PhaseComplete
	PhaseError
	PhaseSaveConfig
)

// Wizard is the main orchestrator for the wizard interface.
type Wizard struct {
	state *WizardState

	// Current phase
	phase Phase

	// Screen instances
	globalScreen     *screens.GlobalScreen
	patientScreen    *screens.PatientScreen
	bulkPatientScreen *screens.BulkPatientScreen
	studyScreen      *screens.StudyScreen
	bulkStudyScreen  *screens.BulkStudyScreen
	seriesScreen     *screens.SeriesScreen
	bulkSeriesScreen *screens.BulkSeriesScreen
	summaryScreen    *screens.SummaryScreen
	progressScreen   *screens.ProgressScreen
	completionScreen *screens.CompletionScreen
	errorScreen      *screens.ErrorScreen

	// Save config form
	saveConfigForm *huh.Form
	configPath     string

	// Tracking indices for iteration
	currentPatientIndex int
	currentStudyIndex   int
	currentSeriesIndex  int

	// Bulk mode flags
	bulkPatients bool // Generate remaining patients automatically
	bulkStudies  bool // Generate remaining studies automatically
	bulkSeries   bool // Generate remaining series automatically

	// Window size
	width  int
	height int

	// Final state
	cancelled bool
	finished  bool
	err       error
}

// NewWizard creates a new wizard with default or loaded state.
func NewWizard(state *WizardState) *Wizard {
	if state == nil {
		state = &WizardState{
			Global: GlobalConfig{
				Modality:          "MR",
				TotalImages:       50,
				TotalSize:         "500MB",
				OutputDir:         "dicom_series",
				NumPatients:       1,
				StudiesPerPatient: 1,
				SeriesPerStudy:    1,
			},
		}
	}

	w := &Wizard{
		state: state,
		phase: PhaseGlobal,
	}

	// Initialize the global screen
	w.globalScreen = screens.NewGlobalScreen(&w.state.Global)

	return w
}

// Init implements tea.Model.
func (w *Wizard) Init() tea.Cmd {
	return w.globalScreen.Init()
}

// Update implements tea.Model.
func (w *Wizard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle window size for all phases
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		w.width = wsm.Width
		w.height = wsm.Height
	}

	switch w.phase {
	case PhaseGlobal:
		return w.updateGlobal(msg)
	case PhasePatient:
		return w.updatePatient(msg)
	case PhaseBulkPatient:
		return w.updateBulkPatient(msg)
	case PhaseStudy:
		return w.updateStudy(msg)
	case PhaseBulkStudy:
		return w.updateBulkStudy(msg)
	case PhaseSeries:
		return w.updateSeries(msg)
	case PhaseBulkSeries:
		return w.updateBulkSeries(msg)
	case PhaseSummary:
		return w.updateSummary(msg)
	case PhaseSaveConfig:
		return w.updateSaveConfig(msg)
	case PhaseProgress:
		return w.updateProgress(msg)
	case PhaseComplete:
		return w.updateComplete(msg)
	case PhaseError:
		return w.updateError(msg)
	}

	return w, nil
}

// View implements tea.Model.
func (w *Wizard) View() string {
	switch w.phase {
	case PhaseGlobal:
		return w.globalScreen.View()
	case PhasePatient:
		return w.patientScreen.View()
	case PhaseBulkPatient:
		return w.bulkPatientScreen.View()
	case PhaseStudy:
		return w.studyScreen.View()
	case PhaseBulkStudy:
		return w.bulkStudyScreen.View()
	case PhaseSeries:
		return w.seriesScreen.View()
	case PhaseBulkSeries:
		return w.bulkSeriesScreen.View()
	case PhaseSummary:
		return w.summaryScreen.View()
	case PhaseSaveConfig:
		return w.viewSaveConfig()
	case PhaseProgress:
		return w.progressScreen.View()
	case PhaseComplete:
		return w.completionScreen.View()
	case PhaseError:
		return w.errorScreen.View()
	}

	return ""
}

// updateGlobal handles updates in the global configuration phase.
func (w *Wizard) updateGlobal(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := w.globalScreen.Update(msg)
	if gs, ok := model.(*screens.GlobalScreen); ok {
		w.globalScreen = gs
	}

	if w.globalScreen.Cancelled() {
		w.cancelled = true
		return w, tea.Quit
	}

	if w.globalScreen.Done() {
		// Initialize patient structure
		w.initializePatients()
		// Move to patient configuration
		w.transitionToPatient(0)
	}

	return w, cmd
}

// initializePatients creates empty patient structures based on global config.
func (w *Wizard) initializePatients() {
	numPatients := w.state.Global.NumPatients
	if numPatients <= 0 {
		numPatients = 1
	}

	w.state.Patients = make([]PatientConfig, numPatients)
	for i := range w.state.Patients {
		// Initialize empty studies for each patient
		studiesPerPatient := w.state.Global.StudiesPerPatient
		if studiesPerPatient <= 0 {
			studiesPerPatient = 1
		}
		w.state.Patients[i].Studies = make([]StudyConfig, studiesPerPatient)

		// Initialize empty series for each study
		seriesPerStudy := w.state.Global.SeriesPerStudy
		if seriesPerStudy <= 0 {
			seriesPerStudy = 1
		}
		for j := range w.state.Patients[i].Studies {
			w.state.Patients[i].Studies[j].Series = make([]SeriesConfig, seriesPerStudy)
		}
	}
}

// transitionToPatient starts patient configuration for the given index.
func (w *Wizard) transitionToPatient(index int) {
	w.currentPatientIndex = index
	w.phase = PhasePatient
	w.patientScreen = screens.NewPatientScreen(
		&w.state.Patients[index],
		index,
		len(w.state.Patients),
	)
}

// updatePatient handles updates in the patient configuration phase.
func (w *Wizard) updatePatient(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := w.patientScreen.Update(msg)
	if ps, ok := model.(*screens.PatientScreen); ok {
		w.patientScreen = ps
	}

	if w.patientScreen.Cancelled() {
		w.cancelled = true
		return w, tea.Quit
	}

	if w.patientScreen.Done() {
		// If accepting defaults for studies, generate them
		if w.patientScreen.AcceptDefaults() {
			w.generateDefaultStudies(w.currentPatientIndex)
		}

		// Check if there are more patients to configure
		if w.currentPatientIndex == 0 && len(w.state.Patients) > 1 {
			// Show bulk patient choice
			w.phase = PhaseBulkPatient
			w.bulkPatientScreen = screens.NewBulkPatientScreen(len(w.state.Patients) - 1)
			return w, w.bulkPatientScreen.Init()
		}

		// Move to study configuration for this patient (if not accepting defaults)
		if !w.patientScreen.AcceptDefaults() {
			w.transitionToStudy(w.currentPatientIndex, 0)
			return w, w.studyScreen.Init()
		}

		// Move to next patient or summary
		return w.advanceToNextPatientOrSummary()
	}

	return w, cmd
}

// updateBulkPatient handles updates in the bulk patient choice phase.
func (w *Wizard) updateBulkPatient(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := w.bulkPatientScreen.Update(msg)
	if bps, ok := model.(*screens.BulkPatientScreen); ok {
		w.bulkPatientScreen = bps
	}

	if w.bulkPatientScreen.Cancelled() {
		w.cancelled = true
		return w, tea.Quit
	}

	if w.bulkPatientScreen.Done() {
		choice := w.bulkPatientScreen.Choice()
		if choice == screens.BulkGenerate {
			// Generate all remaining patients automatically
			w.generateRemainingPatients()
			// Move to study configuration for first patient
			if !w.patientScreen.AcceptDefaults() {
				w.transitionToStudy(0, 0)
				return w, w.studyScreen.Init()
			}
			// Or go to summary if accepting defaults
			return w.transitionToSummary()
		}
		// Configure each patient individually
		w.bulkPatients = false
		w.transitionToPatient(1)
		return w, w.patientScreen.Init()
	}

	return w, cmd
}

// generateRemainingPatients generates default values for patients after the first.
func (w *Wizard) generateRemainingPatients() {
	rng := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0))
	for i := 1; i < len(w.state.Patients); i++ {
		w.state.Patients[i] = generateDefaultPatient(i, rng)
		// Also generate default studies
		w.generateDefaultStudies(i)
	}
}

// generateDefaultPatient creates a patient with random default values.
func generateDefaultPatient(index int, rng *rand.Rand) PatientConfig {
	sex := []string{"M", "F"}[rng.IntN(2)]
	birthYear := 1950 + rng.IntN(50) // 1950-2000
	birthMonth := 1 + rng.IntN(12)
	birthDay := 1 + rng.IntN(28)

	return PatientConfig{
		Name:      util.GeneratePatientName(sex, rng),
		ID:        fmt.Sprintf("PAT%06d", index+1),
		BirthDate: fmt.Sprintf("%04d-%02d-%02d", birthYear, birthMonth, birthDay),
		Sex:       sex,
	}
}

// generateDefaultStudies generates default studies and series for a patient.
func (w *Wizard) generateDefaultStudies(patientIndex int) {
	patient := &w.state.Patients[patientIndex]
	rng := rand.New(rand.NewPCG(uint64(time.Now().UnixNano())+uint64(patientIndex), 0))

	studiesPerPatient := w.state.Global.StudiesPerPatient
	if studiesPerPatient <= 0 {
		studiesPerPatient = 1
	}

	patient.Studies = make([]StudyConfig, studiesPerPatient)
	for i := range patient.Studies {
		patient.Studies[i] = generateDefaultStudy(w.state.Global.Modality, rng)
		// Generate default series
		w.generateDefaultSeries(patientIndex, i)
	}
}

// generateDefaultStudy creates a study with random default values.
func generateDefaultStudy(modality string, rng *rand.Rand) StudyConfig {
	bodyParts := []string{"HEAD", "BRAIN", "CHEST", "ABDOMEN", "SPINE"}
	bodyPart := bodyParts[rng.IntN(len(bodyParts))]

	studyDate := time.Now().AddDate(0, -rng.IntN(12), -rng.IntN(30))

	return StudyConfig{
		Description:     fmt.Sprintf("%s %s", bodyPart, modality),
		Date:            studyDate.Format("2006-01-02"),
		AccessionNumber: fmt.Sprintf("ACC-%06d", rng.IntN(1000000)),
		BodyPart:        bodyPart,
		Priority:        "ROUTINE",
	}
}

// generateDefaultSeries generates default series for a study.
func (w *Wizard) generateDefaultSeries(patientIndex, studyIndex int) {
	study := &w.state.Patients[patientIndex].Studies[studyIndex]

	seriesPerStudy := w.state.Global.SeriesPerStudy
	if seriesPerStudy <= 0 {
		seriesPerStudy = 1
	}

	// Calculate images per series
	totalImages := w.state.Global.TotalImages
	totalStudies := w.state.Global.NumPatients * w.state.Global.StudiesPerPatient
	totalSeries := totalStudies * seriesPerStudy
	imagesPerSeries := totalImages / totalSeries
	if imagesPerSeries < 1 {
		imagesPerSeries = 1
	}

	study.Series = make([]SeriesConfig, seriesPerStudy)
	orientations := []string{"AXIAL", "SAGITTAL", "CORONAL"}
	for i := range study.Series {
		study.Series[i] = SeriesConfig{
			Description: fmt.Sprintf("Series %d", i+1),
			Orientation: orientations[i%len(orientations)],
			ImageCount:  imagesPerSeries,
		}
	}
}

// transitionToStudy starts study configuration for the given patient and study index.
func (w *Wizard) transitionToStudy(patientIndex, studyIndex int) {
	w.currentPatientIndex = patientIndex
	w.currentStudyIndex = studyIndex
	w.phase = PhaseStudy

	patient := &w.state.Patients[patientIndex]
	totalStudies := len(patient.Studies)

	w.studyScreen = screens.NewStudyScreen(
		&patient.Studies[studyIndex],
		studyIndex,
		totalStudies,
		patient.Name,
		w.state.Global.Modality,
	)
}

// updateStudy handles updates in the study configuration phase.
func (w *Wizard) updateStudy(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := w.studyScreen.Update(msg)
	if ss, ok := model.(*screens.StudyScreen); ok {
		w.studyScreen = ss
	}

	if w.studyScreen.Cancelled() {
		w.cancelled = true
		return w, tea.Quit
	}

	if w.studyScreen.Done() {
		patient := &w.state.Patients[w.currentPatientIndex]
		totalStudies := len(patient.Studies)

		// If accepting defaults for series, generate them
		if w.studyScreen.AcceptDefaults() {
			w.generateDefaultSeries(w.currentPatientIndex, w.currentStudyIndex)
		}

		// Check if there are more studies for this patient
		if w.currentStudyIndex == 0 && totalStudies > 1 {
			// Show bulk study choice
			w.phase = PhaseBulkStudy
			w.bulkStudyScreen = screens.NewBulkStudyScreen(
				totalStudies-1,
				patient.Name,
			)
			return w, w.bulkStudyScreen.Init()
		}

		// Move to series configuration for this study (if not accepting defaults)
		if !w.studyScreen.AcceptDefaults() {
			w.transitionToSeries(w.currentPatientIndex, w.currentStudyIndex, 0)
			return w, w.seriesScreen.Init()
		}

		// Move to next study or patient
		return w.advanceToNextStudyOrPatient()
	}

	return w, cmd
}

// updateBulkStudy handles updates in the bulk study choice phase.
func (w *Wizard) updateBulkStudy(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := w.bulkStudyScreen.Update(msg)
	if bss, ok := model.(*screens.BulkStudyScreen); ok {
		w.bulkStudyScreen = bss
	}

	if w.bulkStudyScreen.Cancelled() {
		w.cancelled = true
		return w, tea.Quit
	}

	if w.bulkStudyScreen.Done() {
		choice := w.bulkStudyScreen.Choice()
		if choice == screens.BulkStudyGenerate {
			// Generate all remaining studies automatically for this patient
			w.generateRemainingStudies(w.currentPatientIndex)
			// Move to series configuration for first study (if not accepting defaults)
			if !w.studyScreen.AcceptDefaults() {
				w.transitionToSeries(w.currentPatientIndex, 0, 0)
				return w, w.seriesScreen.Init()
			}
			// Or move to next patient
			return w.advanceToNextPatientOrSummary()
		}
		// Configure each study individually
		w.bulkStudies = false
		w.transitionToStudy(w.currentPatientIndex, 1)
		return w, w.studyScreen.Init()
	}

	return w, cmd
}

// generateRemainingStudies generates default values for studies after the first.
func (w *Wizard) generateRemainingStudies(patientIndex int) {
	patient := &w.state.Patients[patientIndex]
	rng := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0))

	for i := 1; i < len(patient.Studies); i++ {
		patient.Studies[i] = generateDefaultStudy(w.state.Global.Modality, rng)
		w.generateDefaultSeries(patientIndex, i)
	}
}

// transitionToSeries starts series configuration for the given study.
func (w *Wizard) transitionToSeries(patientIndex, studyIndex, seriesIndex int) {
	w.currentPatientIndex = patientIndex
	w.currentStudyIndex = studyIndex
	w.currentSeriesIndex = seriesIndex
	w.phase = PhaseSeries

	study := &w.state.Patients[patientIndex].Studies[studyIndex]
	totalSeries := len(study.Series)

	// Calculate images to distribute
	totalImages := w.state.Global.TotalImages
	totalStudies := w.state.Global.NumPatients * w.state.Global.StudiesPerPatient
	imagesPerStudy := totalImages / totalStudies

	w.seriesScreen = screens.NewSeriesScreen(
		&study.Series[seriesIndex],
		seriesIndex,
		totalSeries,
		study.Description,
		w.state.Global.Modality,
		study.BodyPart,
		imagesPerStudy,
	)
}

// updateSeries handles updates in the series configuration phase.
func (w *Wizard) updateSeries(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := w.seriesScreen.Update(msg)
	if ss, ok := model.(*screens.SeriesScreen); ok {
		w.seriesScreen = ss
	}

	if w.seriesScreen.Cancelled() {
		w.cancelled = true
		return w, tea.Quit
	}

	if w.seriesScreen.Done() {
		study := &w.state.Patients[w.currentPatientIndex].Studies[w.currentStudyIndex]
		totalSeries := len(study.Series)

		// Check if there are more series for this study
		if w.currentSeriesIndex == 0 && totalSeries > 1 {
			// Show bulk series choice
			w.phase = PhaseBulkSeries
			w.bulkSeriesScreen = screens.NewBulkSeriesScreen(
				totalSeries-1,
				study.Description,
			)
			return w, w.bulkSeriesScreen.Init()
		}

		// Move to next series, study, or patient
		return w.advanceToNextSeriesOrStudy()
	}

	return w, cmd
}

// updateBulkSeries handles updates in the bulk series choice phase.
func (w *Wizard) updateBulkSeries(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := w.bulkSeriesScreen.Update(msg)
	if bss, ok := model.(*screens.BulkSeriesScreen); ok {
		w.bulkSeriesScreen = bss
	}

	if w.bulkSeriesScreen.Cancelled() {
		w.cancelled = true
		return w, tea.Quit
	}

	if w.bulkSeriesScreen.Done() {
		choice := w.bulkSeriesScreen.Choice()
		if choice == screens.BulkSeriesGenerate {
			// Generate all remaining series automatically
			w.generateRemainingSeries(w.currentPatientIndex, w.currentStudyIndex)
			// Move to next study or patient
			return w.advanceToNextStudyOrPatient()
		}
		// Configure each series individually
		w.bulkSeries = false
		w.transitionToSeries(w.currentPatientIndex, w.currentStudyIndex, 1)
		return w, w.seriesScreen.Init()
	}

	return w, cmd
}

// generateRemainingSeries generates default values for series after the first.
func (w *Wizard) generateRemainingSeries(patientIndex, studyIndex int) {
	study := &w.state.Patients[patientIndex].Studies[studyIndex]
	orientations := []string{"AXIAL", "SAGITTAL", "CORONAL"}

	// Calculate images per series
	totalImages := w.state.Global.TotalImages
	totalStudies := w.state.Global.NumPatients * w.state.Global.StudiesPerPatient
	totalSeries := totalStudies * len(study.Series)
	imagesPerSeries := totalImages / totalSeries
	if imagesPerSeries < 1 {
		imagesPerSeries = 1
	}

	for i := 1; i < len(study.Series); i++ {
		study.Series[i] = SeriesConfig{
			Description: fmt.Sprintf("Series %d", i+1),
			Orientation: orientations[i%len(orientations)],
			ImageCount:  imagesPerSeries,
		}
	}
}

// advanceToNextSeriesOrStudy moves to the next series or advances to next study.
func (w *Wizard) advanceToNextSeriesOrStudy() (tea.Model, tea.Cmd) {
	study := &w.state.Patients[w.currentPatientIndex].Studies[w.currentStudyIndex]

	// Try next series
	if w.currentSeriesIndex+1 < len(study.Series) {
		w.transitionToSeries(w.currentPatientIndex, w.currentStudyIndex, w.currentSeriesIndex+1)
		return w, w.seriesScreen.Init()
	}

	// Move to next study
	return w.advanceToNextStudyOrPatient()
}

// advanceToNextStudyOrPatient moves to the next study or advances to next patient.
func (w *Wizard) advanceToNextStudyOrPatient() (tea.Model, tea.Cmd) {
	patient := &w.state.Patients[w.currentPatientIndex]

	// Try next study
	if w.currentStudyIndex+1 < len(patient.Studies) {
		w.transitionToStudy(w.currentPatientIndex, w.currentStudyIndex+1)
		return w, w.studyScreen.Init()
	}

	// Move to next patient
	return w.advanceToNextPatientOrSummary()
}

// advanceToNextPatientOrSummary moves to the next patient or shows summary.
func (w *Wizard) advanceToNextPatientOrSummary() (tea.Model, tea.Cmd) {
	// Try next patient
	if w.currentPatientIndex+1 < len(w.state.Patients) {
		w.transitionToPatient(w.currentPatientIndex + 1)
		return w, w.patientScreen.Init()
	}

	// All patients configured, show summary
	return w.transitionToSummary()
}

// transitionToSummary moves to the summary screen.
func (w *Wizard) transitionToSummary() (tea.Model, tea.Cmd) {
	w.phase = PhaseSummary
	w.summaryScreen = screens.NewSummaryScreen(w.state)
	return w, w.summaryScreen.Init()
}

// updateSummary handles updates in the summary phase.
func (w *Wizard) updateSummary(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := w.summaryScreen.Update(msg)
	if ss, ok := model.(*screens.SummaryScreen); ok {
		w.summaryScreen = ss
	}

	if w.summaryScreen.Cancelled() {
		w.cancelled = true
		return w, tea.Quit
	}

	if w.summaryScreen.Done() {
		switch w.summaryScreen.Action() {
		case screens.SummaryActionBack:
			// Go back to the first patient
			w.transitionToPatient(0)
			return w, w.patientScreen.Init()

		case screens.SummaryActionGenerate:
			// Start generation
			return w.startGeneration()

		case screens.SummaryActionSaveConfig:
			// Show save config dialog
			return w.transitionToSaveConfig()

		case screens.SummaryActionCancel:
			w.cancelled = true
			return w, tea.Quit
		}
	}

	return w, cmd
}

// transitionToSaveConfig shows the save config dialog.
func (w *Wizard) transitionToSaveConfig() (tea.Model, tea.Cmd) {
	w.phase = PhaseSaveConfig
	w.configPath = "wizard-config.yaml"

	w.saveConfigForm = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("config_path").
				Title("Save configuration to").
				Description("Enter the path for the YAML config file").
				Value(&w.configPath).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("path is required")
					}
					return nil
				}),
		),
	).WithShowHelp(false)

	return w, w.saveConfigForm.Init()
}

// updateSaveConfig handles updates in the save config phase.
func (w *Wizard) updateSaveConfig(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			// Go back to summary
			return w.transitionToSummary()
		case "ctrl+c":
			w.cancelled = true
			return w, tea.Quit
		}
	}

	form, cmd := w.saveConfigForm.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		w.saveConfigForm = f
	}

	if w.saveConfigForm.State == huh.StateCompleted {
		// Save the config
		if err := SaveToYAML(w.state, w.configPath); err != nil {
			w.err = err
			w.phase = PhaseError
			w.errorScreen = screens.NewErrorScreen(err)
			return w, nil
		}

		// Go back to summary with success message
		return w.transitionToSummary()
	}

	return w, cmd
}

// viewSaveConfig renders the save config dialog.
func (w *Wizard) viewSaveConfig() string {
	title := components.TitleStyle.Render("Save Configuration")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		w.saveConfigForm.View(),
		"",
		"Enter: Save | Esc: Back",
	)

	return content
}

// startGeneration begins the DICOM generation process.
func (w *Wizard) startGeneration() (tea.Model, tea.Cmd) {
	w.phase = PhaseProgress
	w.progressScreen = screens.NewProgressScreen(w.state.Global.TotalImages)

	// Start generation in a goroutine and send progress updates
	return w, func() tea.Msg {
		startTime := time.Now()

		opts, err := w.toGeneratorOptions()
		if err != nil {
			return screens.ErrorMsg{Error: err}
		}

		files, err := dicom.GenerateDICOMSeries(opts)
		if err != nil {
			return screens.ErrorMsg{Error: err}
		}

		// Organize into DICOMDIR structure (PT/ST/SE hierarchy)
		if err := dicom.OrganizeFilesIntoDICOMDIR(opts.OutputDir, files, true); err != nil {
			return screens.ErrorMsg{Error: fmt.Errorf("creating DICOMDIR: %w", err)}
		}

		// Calculate total size from organized files
		var totalSize int64
		filepath.Walk(opts.OutputDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				totalSize += info.Size()
			}
			return nil
		})

		return screens.CompletionMsg{
			TotalFiles: len(files),
			TotalSize:  totalSize,
			Duration:   time.Since(startTime),
			OutputDir:  opts.OutputDir,
		}
	}
}

// toGeneratorOptions converts WizardState to dicom.GeneratorOptions.
func (w *Wizard) toGeneratorOptions() (dicom.GeneratorOptions, error) {
	state := w.state

	// Parse modality
	modality := modalities.Modality(state.Global.Modality)
	if !modalities.IsValid(state.Global.Modality) {
		modality = modalities.MR // Default to MR
	}

	// Calculate total studies
	totalStudies := 0
	for _, patient := range state.Patients {
		totalStudies += len(patient.Studies)
	}
	if totalStudies == 0 {
		totalStudies = state.Global.NumPatients * state.Global.StudiesPerPatient
	}

	// Build study descriptions
	var studyDescriptions []string
	for _, patient := range state.Patients {
		for _, study := range patient.Studies {
			studyDescriptions = append(studyDescriptions, study.Description)
		}
	}

	// Aggregate custom tags from all studies and series
	customTags := make(util.ParsedTags)
	for _, patient := range state.Patients {
		for _, study := range patient.Studies {
			for k, v := range study.CustomTags {
				customTags[k] = v
			}
			for _, series := range study.Series {
				for k, v := range series.CustomTags {
					customTags[k] = v
				}
			}
		}
	}

	// Determine series per study
	seriesPerStudy := state.Global.SeriesPerStudy
	if seriesPerStudy <= 0 {
		seriesPerStudy = 1
	}

	opts := dicom.GeneratorOptions{
		NumImages:         state.Global.TotalImages,
		TotalSize:         state.Global.TotalSize,
		OutputDir:         state.Global.OutputDir,
		Seed:              state.Global.Seed,
		NumStudies:        totalStudies,
		NumPatients:       state.Global.NumPatients,
		Modality:          modality,
		SeriesPerStudy:    util.SeriesRange{Min: seriesPerStudy, Max: seriesPerStudy},
		StudyDescriptions: studyDescriptions,
		CustomTags:        customTags,
		Quiet:             true, // Suppress output for TUI integration
	}

	// Extract body part from first study if available
	if len(state.Patients) > 0 && len(state.Patients[0].Studies) > 0 {
		opts.BodyPart = state.Patients[0].Studies[0].BodyPart
		opts.Institution = state.Patients[0].Studies[0].Institution
		opts.Department = state.Patients[0].Studies[0].Department
	}

	return opts, nil
}

// updateProgress handles updates in the progress phase.
func (w *Wizard) updateProgress(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case screens.ProgressMsg:
		w.progressScreen.SetProgress(msg.Current, msg.Total, msg.Path)
		return w, nil

	case screens.CompletionMsg:
		w.phase = PhaseComplete
		w.completionScreen = screens.NewCompletionScreen(msg)
		return w, nil

	case screens.ErrorMsg:
		w.phase = PhaseError
		w.err = msg.Error
		w.errorScreen = screens.NewErrorScreen(msg.Error)
		return w, nil
	}

	model, cmd := w.progressScreen.Update(msg)
	if ps, ok := model.(*screens.ProgressScreen); ok {
		w.progressScreen = ps
	}

	if w.progressScreen.Cancelled() {
		w.cancelled = true
		return w, tea.Quit
	}

	return w, cmd
}

// updateComplete handles updates in the completion phase.
func (w *Wizard) updateComplete(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := w.completionScreen.Update(msg)
	if cs, ok := model.(*screens.CompletionScreen); ok {
		w.completionScreen = cs
	}

	if w.completionScreen.Done() {
		w.finished = true
		return w, tea.Quit
	}

	return w, cmd
}

// updateError handles updates in the error phase.
func (w *Wizard) updateError(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := w.errorScreen.Update(msg)
	if es, ok := model.(*screens.ErrorScreen); ok {
		w.errorScreen = es
	}

	if w.errorScreen.Done() {
		w.finished = true
		return w, tea.Quit
	}

	return w, cmd
}

// Run starts the interactive wizard for DICOM generation configuration.
// If fromConfig is provided, it loads the configuration from that YAML file.
func Run(fromConfig string) error {
	var state *WizardState

	// Load config if provided
	if fromConfig != "" {
		absPath, err := filepath.Abs(fromConfig)
		if err != nil {
			return fmt.Errorf("resolving config path: %w", err)
		}

		loaded, err := LoadFromYAML(absPath)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		state = loaded
	}

	// Create and run the wizard
	wizard := NewWizard(state)
	p := tea.NewProgram(wizard, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("running wizard: %w", err)
	}

	// Check final state
	if w, ok := finalModel.(*Wizard); ok {
		if w.cancelled {
			return nil // User cancelled, not an error
		}
		if w.err != nil {
			return w.err
		}
	}

	return nil
}
