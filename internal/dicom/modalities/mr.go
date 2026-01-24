package modalities

import (
	"math/rand/v2"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

// MRGenerator generates MR (Magnetic Resonance) specific metadata.
type MRGenerator struct{}

// Modality returns the MR modality type.
func (g *MRGenerator) Modality() Modality {
	return MR
}

// SOPClassUID returns the MR Image Storage SOP Class UID.
func (g *MRGenerator) SOPClassUID() string {
	return "1.2.840.10008.5.1.4.1.1.4"
}

// Scanners returns available MR scanner configurations.
func (g *MRGenerator) Scanners() []Scanner {
	return []Scanner{
		{Manufacturer: "SIEMENS", Model: "Avanto", FieldStrength: 1.5},
		{Manufacturer: "SIEMENS", Model: "Skyra", FieldStrength: 3.0},
		{Manufacturer: "GE MEDICAL SYSTEMS", Model: "Signa HDxt", FieldStrength: 1.5},
		{Manufacturer: "GE MEDICAL SYSTEMS", Model: "Discovery MR750", FieldStrength: 3.0},
		{Manufacturer: "PHILIPS", Model: "Achieva", FieldStrength: 1.5},
		{Manufacturer: "PHILIPS", Model: "Ingenia", FieldStrength: 3.0},
	}
}

// GenerateSeriesParams generates MR-specific parameters for a series.
func (g *MRGenerator) GenerateSeriesParams(scanner Scanner, rng *rand.Rand) SeriesParams {
	sequences := []string{"T1_MPRAGE", "T1_SE", "T2_FSE", "T2_FLAIR"}

	params := SeriesParams{
		Modality:              MR,
		Scanner:               scanner,
		PixelSpacing:          0.5 + rng.Float64()*1.5,  // 0.5-2.0 mm
		SliceThickness:        1.0 + rng.Float64()*4.0,  // 1.0-5.0 mm
		EchoTime:              10.0 + rng.Float64()*20.0, // 10-30 ms
		RepetitionTime:        400.0 + rng.Float64()*400.0, // 400-800 ms
		FlipAngle:             60.0 + rng.Float64()*30.0, // 60-90 degrees
		SequenceName:          sequences[rng.IntN(len(sequences))],
		MagneticFieldStrength: scanner.FieldStrength,
		ImagingFrequency:      scanner.FieldStrength * 42.58, // MHz
		WindowCenter:          500.0 + rng.Float64()*1000.0, // 500-1500
		WindowWidth:           1000.0 + rng.Float64()*1000.0, // 1000-2000
	}
	params.SpacingBetweenSlices = params.SliceThickness + rng.Float64()*0.5

	return params
}

// PixelConfig returns MR pixel data configuration.
func (g *MRGenerator) PixelConfig() PixelConfig {
	return PixelConfig{
		BitsAllocated:       16,
		BitsStored:          12,
		HighBit:             11,
		PixelRepresentation: 0, // Unsigned
		MinValue:            0,
		MaxValue:            4095,
		BaseValue:           2048,
	}
}

// AppendModalityElements appends MR-specific DICOM elements to a dataset.
func (g *MRGenerator) AppendModalityElements(ds *dicom.Dataset, params SeriesParams) error {
	elements := []*dicom.Element{
		mustNewElement(tag.MagneticFieldStrength, []string{floatToDS(params.MagneticFieldStrength)}),
		mustNewElement(tag.ImagingFrequency, []string{floatToDS(params.ImagingFrequency)}),
	}

	if params.EchoTime != 0 {
		elements = append(elements, mustNewElement(tag.EchoTime, []string{floatToDS(params.EchoTime)}))
	}
	if params.RepetitionTime != 0 {
		elements = append(elements, mustNewElement(tag.RepetitionTime, []string{floatToDS(params.RepetitionTime)}))
	}
	if params.FlipAngle != 0 {
		elements = append(elements, mustNewElement(tag.FlipAngle, []string{floatToDS(params.FlipAngle)}))
	}
	if params.SequenceName != "" {
		elements = append(elements, mustNewElement(tag.SequenceName, []string{params.SequenceName}))
	}

	ds.Elements = append(ds.Elements, elements...)
	return nil
}

// WindowPresets returns MR window presets.
func (g *MRGenerator) WindowPresets() []WindowPreset {
	return []WindowPreset{
		{Name: "DEFAULT", Center: 500, Width: 1000},
		{Name: "BRIGHT", Center: 300, Width: 600},
		{Name: "CONTRAST", Center: 600, Width: 1200},
	}
}
