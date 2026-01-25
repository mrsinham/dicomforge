// Package wizard provides an interactive TUI for configuring DICOM generation.
package wizard

// WizardState holds the complete state for the wizard interface.
type WizardState struct {
	Global   GlobalConfig
	Patients []PatientConfig
}

// GlobalConfig holds global settings that apply to the entire generation.
type GlobalConfig struct {
	Modality          string
	TotalImages       int
	TotalSize         string
	OutputDir         string
	Seed              int64
	NumPatients       int
	StudiesPerPatient int
	SeriesPerStudy    int
}

// PatientConfig holds configuration for a single patient.
type PatientConfig struct {
	Name      string
	ID        string
	BirthDate string
	Sex       string
	Studies   []StudyConfig
}

// StudyConfig holds configuration for a single study.
type StudyConfig struct {
	Description        string
	Date               string
	AccessionNumber    string
	Institution        string
	Department         string
	BodyPart           string
	Priority           string
	ReferringPhysician string
	CustomTags         map[string]string
	Series             []SeriesConfig
}

// SeriesConfig holds configuration for a single series.
type SeriesConfig struct {
	Description string
	Protocol    string
	Orientation string
	ImageCount  int
	CustomTags  map[string]string
}
