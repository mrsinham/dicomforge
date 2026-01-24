package modalities

import (
	"math/rand/v2"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

// CTGenerator generates CT (Computed Tomography) specific metadata.
type CTGenerator struct{}

// Modality returns the CT modality type.
func (g *CTGenerator) Modality() Modality {
	return CT
}

// SOPClassUID returns the CT Image Storage SOP Class UID.
func (g *CTGenerator) SOPClassUID() string {
	return "1.2.840.10008.5.1.4.1.1.2"
}

// Scanners returns available CT scanner configurations.
func (g *CTGenerator) Scanners() []Scanner {
	return []Scanner{
		{Manufacturer: "SIEMENS", Model: "SOMATOM Definition AS+", DetectorRows: 128},
		{Manufacturer: "SIEMENS", Model: "SOMATOM Force", DetectorRows: 192},
		{Manufacturer: "GE MEDICAL SYSTEMS", Model: "Revolution CT", DetectorRows: 256},
		{Manufacturer: "GE MEDICAL SYSTEMS", Model: "LightSpeed VCT", DetectorRows: 64},
		{Manufacturer: "PHILIPS", Model: "Brilliance iCT", DetectorRows: 256},
		{Manufacturer: "PHILIPS", Model: "Ingenuity CT", DetectorRows: 128},
		{Manufacturer: "CANON", Model: "Aquilion ONE", DetectorRows: 320},
		{Manufacturer: "CANON", Model: "Aquilion Prime", DetectorRows: 80},
	}
}

// GenerateSeriesParams generates CT-specific parameters for a series.
func (g *CTGenerator) GenerateSeriesParams(scanner Scanner, rng *rand.Rand) SeriesParams {
	// KVP options: 80, 100, 120, 140
	kvpOptions := []float64{80, 100, 120, 140}
	kvp := kvpOptions[rng.IntN(len(kvpOptions))]

	// Convolution kernels
	kernels := []string{"SOFT", "STANDARD", "BONE", "LUNG"}
	kernel := kernels[rng.IntN(len(kernels))]

	// Select window preset based on kernel
	var windowCenter, windowWidth float64
	switch kernel {
	case "BONE":
		windowCenter = 400
		windowWidth = 2000
	case "LUNG":
		windowCenter = -600
		windowWidth = 1500
	default: // SOFT, STANDARD
		windowCenter = 40
		windowWidth = 400
	}

	params := SeriesParams{
		Modality:             CT,
		Scanner:              scanner,
		PixelSpacing:         0.5 + rng.Float64()*0.5, // 0.5-1.0 mm
		SliceThickness:       0.5 + rng.Float64()*2.5, // 0.5-3.0 mm
		KVP:                  kvp,
		XRayTubeCurrent:      100 + rng.IntN(301), // 100-400 mA
		ConvolutionKernel:    kernel,
		RescaleIntercept:     -1024, // Standard CT offset for HU
		RescaleSlope:         1,     // Standard CT scale
		GantryTilt:           0,     // Usually 0 for modern CT
		WindowCenter:         windowCenter,
		WindowWidth:          windowWidth,
	}
	params.SpacingBetweenSlices = params.SliceThickness

	return params
}

// PixelConfig returns CT pixel data configuration.
func (g *CTGenerator) PixelConfig() PixelConfig {
	return PixelConfig{
		BitsAllocated:       16,
		BitsStored:          16,
		HighBit:             15,
		PixelRepresentation: 1, // Signed (for Hounsfield units)
		MinValue:            -1024, // Air in HU (after rescale)
		MaxValue:            3071,  // Dense bone in HU (after rescale)
		BaseValue:           1024,  // Water = 0 HU (stored as 1024 with -1024 intercept)
	}
}

// AppendModalityElements appends CT-specific DICOM elements to a dataset.
func (g *CTGenerator) AppendModalityElements(ds *dicom.Dataset, params SeriesParams) error {
	elements := []*dicom.Element{
		mustNewElement(tag.KVP, []string{floatToDS(params.KVP)}),
		mustNewElement(tag.XRayTubeCurrent, []string{intToIS(params.XRayTubeCurrent)}),
		mustNewElement(tag.ConvolutionKernel, []string{params.ConvolutionKernel}),
		mustNewElement(tag.RescaleIntercept, []string{floatToDS(params.RescaleIntercept)}),
		mustNewElement(tag.RescaleSlope, []string{floatToDS(params.RescaleSlope)}),
		mustNewElement(tag.RescaleType, []string{"HU"}),
		mustNewElement(tag.GantryDetectorTilt, []string{floatToDS(params.GantryTilt)}),
	}

	ds.Elements = append(ds.Elements, elements...)
	return nil
}

// WindowPresets returns CT window presets.
func (g *CTGenerator) WindowPresets() []WindowPreset {
	return []WindowPreset{
		{Name: "BRAIN", Center: 40, Width: 80},
		{Name: "SUBDURAL", Center: 75, Width: 215},
		{Name: "BONE", Center: 400, Width: 2000},
		{Name: "LUNG", Center: -600, Width: 1500},
		{Name: "MEDIASTINUM", Center: 40, Width: 400},
		{Name: "ABDOMEN", Center: 40, Width: 350},
		{Name: "LIVER", Center: 60, Width: 150},
	}
}
