// Package modalities provides DICOM modality-specific metadata generators.
package modalities

import (
	"math/rand/v2"

	"github.com/suyashkumar/dicom"
)

// Modality represents a DICOM imaging modality type.
type Modality string

const (
	MR Modality = "MR" // Magnetic Resonance
	CT Modality = "CT" // Computed Tomography
)

// AllModalities returns all supported modalities.
func AllModalities() []Modality {
	return []Modality{MR, CT}
}

// IsValid checks if a modality string is valid.
func IsValid(m string) bool {
	for _, valid := range AllModalities() {
		if string(valid) == m {
			return true
		}
	}
	return false
}

// Scanner represents an imaging device configuration.
type Scanner struct {
	Manufacturer string
	Model        string
	// MR-specific
	FieldStrength float64 // Tesla (1.5, 3.0)
	// CT-specific
	DetectorRows int // Number of detector rows (16, 64, 128, 256)
}

// SeriesParams holds modality-specific parameters for a series.
type SeriesParams struct {
	// Common
	Modality     Modality
	Scanner      Scanner
	WindowCenter float64
	WindowWidth  float64

	// MR-specific
	EchoTime             float64
	RepetitionTime       float64
	FlipAngle            float64
	SequenceName         string
	MagneticFieldStrength float64
	ImagingFrequency     float64

	// CT-specific
	KVP                float64 // Tube voltage (kV)
	XRayTubeCurrent    int     // Tube current (mA)
	ConvolutionKernel  string  // Reconstruction kernel
	RescaleIntercept   float64 // HU offset (-1024)
	RescaleSlope       float64 // HU scale (1)
	GantryTilt         float64 // Gantry tilt angle

	// Geometry (common)
	PixelSpacing         float64
	SliceThickness       float64
	SpacingBetweenSlices float64
}

// PixelConfig holds pixel data configuration for a modality.
type PixelConfig struct {
	BitsAllocated       uint16
	BitsStored          uint16
	HighBit             uint16
	PixelRepresentation uint16 // 0 = unsigned, 1 = signed
	MinValue            int    // Minimum pixel value
	MaxValue            int    // Maximum pixel value
	BaseValue           int    // Base value for synthetic images
}

// Generator defines the interface for modality-specific generators.
type Generator interface {
	// Modality returns the modality type.
	Modality() Modality

	// SOPClassUID returns the SOP Class UID for this modality.
	SOPClassUID() string

	// Scanners returns available scanner configurations.
	Scanners() []Scanner

	// GenerateSeriesParams generates modality-specific parameters for a series.
	GenerateSeriesParams(scanner Scanner, rng *rand.Rand) SeriesParams

	// PixelConfig returns pixel data configuration.
	PixelConfig() PixelConfig

	// AppendModalityElements appends modality-specific DICOM elements to a dataset.
	AppendModalityElements(ds *dicom.Dataset, params SeriesParams) error

	// WindowPresets returns default window presets for this modality.
	WindowPresets() []WindowPreset
}

// WindowPreset represents a window/level preset.
type WindowPreset struct {
	Name   string
	Center float64
	Width  float64
}

// GetGenerator returns the generator for the specified modality.
func GetGenerator(m Modality) Generator {
	switch m {
	case CT:
		return &CTGenerator{}
	case MR:
		fallthrough
	default:
		return &MRGenerator{}
	}
}
