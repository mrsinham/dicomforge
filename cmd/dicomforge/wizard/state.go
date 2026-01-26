// Package wizard provides an interactive TUI for configuring DICOM generation.
package wizard

import "github.com/mrsinham/dicomforge/cmd/dicomforge/wizard/types"

// Re-export types for backwards compatibility
type (
	WizardState   = types.WizardState
	GlobalConfig  = types.GlobalConfig
	PatientConfig = types.PatientConfig
	StudyConfig   = types.StudyConfig
	SeriesConfig  = types.SeriesConfig
)
