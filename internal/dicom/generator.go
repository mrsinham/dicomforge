package dicom

import (
	"fmt"
	"hash/fnv"
	"image"
	"image/color"
	"math"
	randv2 "math/rand/v2"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"sort"

	"github.com/mrsinham/dicomforge/internal/dicom/corruption"
	"github.com/mrsinham/dicomforge/internal/dicom/edgecases"
	"github.com/mrsinham/dicomforge/internal/dicom/modalities"
	"github.com/mrsinham/dicomforge/internal/util"
	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/frame"
	"github.com/suyashkumar/dicom/pkg/tag"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// writeDatasetToFile writes a DICOM dataset to a file
func writeDatasetToFile(filename string, ds dicom.Dataset, opts ...dicom.WriteOption) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return dicom.Write(f, ds, opts...)
}

// drawTextOnFrame16 draws large text overlay on a uint16 frame
func drawTextOnFrame16(nativeFrame *frame.NativeFrame[uint16], width, height int, text string) {
	// Create an RGBA image for drawing (easier to draw text)
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Copy pixel data to RGBA image (convert uint16 to uint8 for display)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			val := nativeFrame.RawData[y*width+x]
			// Scale from uint16 (0-65535) to uint8 (0-255) for drawing
			gray := uint8(val >> 8)
			img.Set(x, y, color.RGBA{gray, gray, gray, 255})
		}
	}

	// Step 1: Render text at base size
	face := basicfont.Face7x13
	baseTextWidth := font.MeasureString(face, text).Ceil()
	baseTextHeight := 13

	// Create a small image for the base text
	textImg := image.NewRGBA(image.Rect(0, 0, baseTextWidth, baseTextHeight))

	// Draw text on the small image (white on transparent)
	drawer := &font.Drawer{
		Dst:  textImg,
		Src:  image.NewUniform(color.RGBA{255, 255, 255, 255}),
		Face: face,
		Dot:  fixed.Point26_6{Y: fixed.I(13)}, // Baseline at height
	}
	drawer.DrawString(text)

	// Step 2: Calculate scale factor to make text 30% of image width
	targetWidth := int(float64(width) * 0.3)
	scaleFactor := float64(targetWidth) / float64(baseTextWidth)

	// Ensure minimum scale for readability
	if scaleFactor < 2.0 {
		scaleFactor = 2.0
	}

	scaledWidth := int(float64(baseTextWidth) * scaleFactor)
	scaledHeight := int(float64(baseTextHeight) * scaleFactor)

	// Step 3: Create scaled text image
	scaledTextImg := image.NewRGBA(image.Rect(0, 0, scaledWidth, scaledHeight))

	// Scale up the text using bilinear interpolation
	draw.BiLinear.Scale(scaledTextImg, scaledTextImg.Bounds(), textImg, textImg.Bounds(), draw.Over, nil)

	// Step 4: Position the text - centered horizontally and vertically
	x := (width - scaledWidth) / 2
	y := (height - scaledHeight) / 2

	// Step 5: Draw thick black outline for visibility
	outlineThickness := max(3, scaledHeight/10) // Proportional outline

	for dx := -outlineThickness; dx <= outlineThickness; dx++ {
		for dy := -outlineThickness; dy <= outlineThickness; dy++ {
			if dx*dx+dy*dy <= outlineThickness*outlineThickness { // Circular outline
				// Draw outline by copying with black color
				for sy := 0; sy < scaledHeight; sy++ {
					for sx := 0; sx < scaledWidth; sx++ {
						r, g, b, a := scaledTextImg.At(sx, sy).RGBA()
						if a > 0 { // If there's text here
							destX := x + sx + dx
							destY := y + sy + dy
							if destX >= 0 && destX < width && destY >= 0 && destY < height {
								// Draw black outline
								img.Set(destX, destY, color.RGBA{0, 0, 0, 255})
							}
						}
						_ = r
						_ = g
						_ = b
					}
				}
			}
		}
	}

	// Step 6: Draw main text (white) on top
	for sy := 0; sy < scaledHeight; sy++ {
		for sx := 0; sx < scaledWidth; sx++ {
			r, g, b, a := scaledTextImg.At(sx, sy).RGBA()
			if a > 0 { // If there's text here
				destX := x + sx
				destY := y + sy
				if destX >= 0 && destX < width && destY >= 0 && destY < height {
					// Blend white text on top
					brightness := (r + g + b) / 3 / 256 // 0-255 range
					img.Set(destX, destY, color.RGBA{uint8(brightness), uint8(brightness), uint8(brightness), 255})
				}
			}
		}
	}

	// Convert back to uint16 and update the frame
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// Average RGB to grayscale, scale back to uint16
			gray := (r + g + b) / 3
			// Scale from 16-bit color space (0-65535) to uint16
			nativeFrame.RawData[y*width+x] = uint16(gray)
		}
	}
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// drawTextOnFrame8 draws large text overlay on a uint8 frame
func drawTextOnFrame8(nativeFrame *frame.NativeFrame[uint8], width, height int, text string) {
	// Create an RGBA image for drawing (easier to draw text)
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Copy pixel data to RGBA image
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			val := nativeFrame.RawData[y*width+x]
			img.Set(x, y, color.RGBA{val, val, val, 255})
		}
	}

	// Step 1: Render text at base size
	face := basicfont.Face7x13
	baseTextWidth := font.MeasureString(face, text).Ceil()
	baseTextHeight := 13

	// Create a small image for the base text
	textImg := image.NewRGBA(image.Rect(0, 0, baseTextWidth, baseTextHeight))

	// Draw text on the small image (white on transparent)
	drawer := &font.Drawer{
		Dst:  textImg,
		Src:  image.NewUniform(color.RGBA{255, 255, 255, 255}),
		Face: face,
		Dot:  fixed.Point26_6{Y: fixed.I(13)}, // Baseline at height
	}
	drawer.DrawString(text)

	// Step 2: Calculate scale factor to make text 30% of image width
	targetWidth := int(float64(width) * 0.3)
	scaleFactor := float64(targetWidth) / float64(baseTextWidth)

	// Ensure minimum scale for readability
	if scaleFactor < 2.0 {
		scaleFactor = 2.0
	}

	scaledWidth := int(float64(baseTextWidth) * scaleFactor)
	scaledHeight := int(float64(baseTextHeight) * scaleFactor)

	// Step 3: Create scaled text image
	scaledTextImg := image.NewRGBA(image.Rect(0, 0, scaledWidth, scaledHeight))

	// Scale up the text using bilinear interpolation
	draw.BiLinear.Scale(scaledTextImg, scaledTextImg.Bounds(), textImg, textImg.Bounds(), draw.Over, nil)

	// Step 4: Position the text - centered horizontally and vertically
	posX := (width - scaledWidth) / 2
	posY := (height - scaledHeight) / 2

	// Step 5: Draw thick black outline for visibility
	outlineThickness := max(3, scaledHeight/10) // Proportional outline

	for dx := -outlineThickness; dx <= outlineThickness; dx++ {
		for dy := -outlineThickness; dy <= outlineThickness; dy++ {
			if dx*dx+dy*dy <= outlineThickness*outlineThickness { // Circular outline
				// Draw outline by copying with black color
				for sy := 0; sy < scaledHeight; sy++ {
					for sx := 0; sx < scaledWidth; sx++ {
						_, _, _, a := scaledTextImg.At(sx, sy).RGBA()
						if a > 0 { // If there's text here
							destX := posX + sx + dx
							destY := posY + sy + dy
							if destX >= 0 && destX < width && destY >= 0 && destY < height {
								// Draw black outline
								img.Set(destX, destY, color.RGBA{0, 0, 0, 255})
							}
						}
					}
				}
			}
		}
	}

	// Step 6: Draw main text (white) on top
	for sy := 0; sy < scaledHeight; sy++ {
		for sx := 0; sx < scaledWidth; sx++ {
			r, g, b, a := scaledTextImg.At(sx, sy).RGBA()
			if a > 0 { // If there's text here
				destX := posX + sx
				destY := posY + sy
				if destX >= 0 && destX < width && destY >= 0 && destY < height {
					// Blend white text on top
					brightness := (r + g + b) / 3 / 256 // 0-255 range
					img.Set(destX, destY, color.RGBA{uint8(brightness), uint8(brightness), uint8(brightness), 255})
				}
			}
		}
	}

	// Convert back to uint8 and update the frame
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// Average RGB to grayscale
			gray := (r + g + b) / 3 / 256 // Scale to 0-255
			nativeFrame.RawData[y*width+x] = uint8(gray)
		}
	}
}

// GeneratorOptions contains all parameters needed to generate a DICOM series
type GeneratorOptions struct {
	NumImages   int
	TotalSize   string
	OutputDir   string
	Seed        int64
	NumStudies  int
	NumPatients int // Number of patients (studies are distributed among patients)
	Workers     int // Number of parallel workers (0 = auto-detect based on CPU cores)

	// Modality selection
	Modality modalities.Modality // Imaging modality (MR, CT, etc.)

	// Multi-series support
	SeriesPerStudy    util.SeriesRange // Number of series per study (default: 1)
	StudyDescriptions []string         // Custom study descriptions (one per study, or empty for auto-generate)

	// Categorization options
	Institution    string        // Fixed institution name (empty = random)
	Department     string        // Fixed department name (empty = random)
	BodyPart       string        // Fixed body part (empty = random per modality)
	Priority       util.Priority // Exam priority
	VariedMetadata bool          // Generate varied institutions/physicians per study

	// Custom tag overrides
	CustomTags util.ParsedTags // User-defined tag overrides

	// Edge case generation
	EdgeCaseConfig edgecases.Config // Edge case generation config

	// Corruption generation (vendor-specific private tags and malformed elements)
	CorruptionConfig corruption.Config

	// Output control
	Quiet            bool                    // Suppress progress output (for TUI integration)
	ProgressCallback func(current, total int) // Optional callback for progress updates

	// Pre-defined patient data (from config file)
	// When set, overrides random generation for patient/study/series metadata
	PredefinedPatients []PredefinedPatient
}

// PredefinedPatient holds pre-configured patient data from config file.
type PredefinedPatient struct {
	Name      string
	ID        string
	BirthDate string
	Sex       string
	Studies   []PredefinedStudy
}

// PredefinedStudy holds pre-configured study data from config file.
type PredefinedStudy struct {
	Description        string
	Date               string
	AccessionNumber    string
	Institution        string
	Department         string
	BodyPart           string
	Priority           string
	ReferringPhysician string
	Series             []PredefinedSeries
}

// PredefinedSeries holds pre-configured series data from config file.
type PredefinedSeries struct {
	Description string
	Protocol    string
	Orientation string
	ImageCount  int // 0 = auto-distribute
}

// getTagValue returns the custom tag value if set, otherwise returns the generated value.
func getTagValue(customTags util.ParsedTags, name, generated string) string {
	if val, ok := customTags.Get(name); ok {
		return val
	}
	return generated
}

// patientInfo holds generated patient data
type patientInfo struct {
	ID        string
	Name      string
	Sex       string
	BirthDate string
}

// imageTask contains all data needed to generate a single DICOM image
type imageTask struct {
	globalIndex      int
	instanceInStudy  int
	instanceInSeries int
	seriesNumber     int
	width            int
	height           int
	filePath         string
	textOverlay      string
	pixelSeed          uint64 // Deterministic seed for this image's pixel generation
	metadata           []*dicom.Element
	pixelConfig        modalities.PixelConfig // Modality-specific pixel configuration
	writeOpts          []dicom.WriteOption    // Write options (e.g., SkipVRVerification for corruption)
	hasMalformedLengths bool                  // Whether to apply malformed length post-processing
	// Result info
	studyUID       string
	seriesUID      string
	sopInstanceUID string
	patientID      string
	studyID        string
}

// GeneratedFile contains information about a generated DICOM file
type GeneratedFile struct {
	Path             string
	StudyUID         string
	SeriesUID        string
	SOPInstanceUID   string
	PatientID        string
	StudyID          string
	SeriesNumber     int
	InstanceNumber   int // Instance number in series
	InstanceInStudy  int // Instance number in study (for backwards compatibility)
}

// generateImageFromTask generates a single DICOM image from a pre-computed task
func generateImageFromTask(task imageTask) error {
	width, height := task.width, task.height
	pixelsPerFrame := width * height
	cfg := task.pixelConfig

	// Create deterministic RNG for this specific image
	rng := randv2.New(randv2.NewPCG(task.pixelSeed, task.pixelSeed))

	// Calculate value range based on pixel config
	valueRange := float64(cfg.MaxValue - cfg.MinValue)
	baseValue := float64(cfg.BaseValue)
	centerX, centerY := float64(width)/2, float64(height)/2
	maxDist := math.Sqrt(centerX*centerX + centerY*centerY)

	// Generate pixel data based on BitsAllocated
	var pixelDataInfo dicom.PixelDataInfo

	if cfg.BitsAllocated == 8 {
		// 8-bit pixel data (e.g., Ultrasound)
		nativeFrame := frame.NewNativeFrame[uint8](8, height, width, pixelsPerFrame, 1)

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				dx := float64(x) - centerX
				dy := float64(y) - centerY
				dist := math.Sqrt(dx*dx + dy*dy)

				normalizedDist := dist / maxDist
				baseIntensity := baseValue + (1.0-normalizedDist)*valueRange*0.3

				largeNoise := (rng.Float64() - 0.5) * valueRange * 0.3
				mediumNoise := (rng.Float64() - 0.5) * valueRange * 0.15
				fineNoise := (rng.Float64() - 0.5) * valueRange * 0.075

				totalNoise := largeNoise + mediumNoise + fineNoise
				intensity := baseIntensity + totalNoise

				minVal := float64(0)
				maxValInt := (1 << cfg.BitsStored) - 1
				maxVal := float64(maxValInt)
				clampedValue := math.Max(minVal, math.Min(maxVal, intensity))
				nativeFrame.RawData[y*width+x] = uint8(clampedValue)
			}
		}

		drawTextOnFrame8(nativeFrame, width, height, task.textOverlay)

		pixelDataInfo = dicom.PixelDataInfo{
			Frames: []*frame.Frame{
				{
					Encapsulated: false,
					NativeData:   nativeFrame,
				},
			},
		}
	} else {
		// 16-bit pixel data (MR, CT, CR, DX, MG)
		nativeFrame := frame.NewNativeFrame[uint16](16, height, width, pixelsPerFrame, 1)

		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				dx := float64(x) - centerX
				dy := float64(y) - centerY
				dist := math.Sqrt(dx*dx + dy*dy)

				normalizedDist := dist / maxDist
				baseIntensity := baseValue + (1.0-normalizedDist)*valueRange*0.3

				largeNoise := (rng.Float64() - 0.5) * valueRange * 0.3
				mediumNoise := (rng.Float64() - 0.5) * valueRange * 0.15
				fineNoise := (rng.Float64() - 0.5) * valueRange * 0.075

				totalNoise := largeNoise + mediumNoise + fineNoise
				intensity := baseIntensity + totalNoise

				minVal := float64(0)
				maxValInt := (1 << cfg.BitsStored) - 1
				maxVal := float64(maxValInt)
				clampedValue := math.Max(minVal, math.Min(maxVal, intensity))
				nativeFrame.RawData[y*width+x] = uint16(clampedValue)
			}
		}

		drawTextOnFrame16(nativeFrame, width, height, task.textOverlay)

		pixelDataInfo = dicom.PixelDataInfo{
			Frames: []*frame.Frame{
				{
					Encapsulated: false,
					NativeData:   nativeFrame,
				},
			},
		}
	}

	// Build complete metadata with pixel data
	elements := make([]*dicom.Element, len(task.metadata)+1)
	copy(elements, task.metadata)
	elements[len(task.metadata)] = mustNewElement(tag.PixelData, pixelDataInfo)

	// Write DICOM file
	if err := writeDatasetToFile(task.filePath, dicom.Dataset{Elements: elements}, task.writeOpts...); err != nil {
		return err
	}

	// Apply malformed length post-processing if needed
	if task.hasMalformedLengths {
		if err := corruption.PatchMalformedLengths(task.filePath); err != nil {
			return fmt.Errorf("patch malformed lengths: %w", err)
		}
	}

	return nil
}

// CalculateDimensions calculates optimal image dimensions based on total size and number of images
func CalculateDimensions(totalBytes int64, numImages int) (width, height int, err error) {
	if totalBytes <= 0 {
		return 0, 0, fmt.Errorf("total bytes must be > 0")
	}
	if numImages <= 0 {
		return 0, 0, fmt.Errorf("number of images must be > 0")
	}

	// Subtract metadata overhead (100KB estimate)
	metadataOverhead := int64(100 * 1024)
	availableBytes := totalBytes - metadataOverhead
	if availableBytes <= 0 {
		return 0, 0, fmt.Errorf("total size too small (need at least 100KB for metadata)")
	}

	// DICOM max size check (2^32 - 10MB ≈ 4.28GB)
	maxDICOMSize := int64(math.Pow(2, 32)) - 10*1024*1024
	if availableBytes > maxDICOMSize {
		availableBytes = maxDICOMSize
	}

	// Calculate total pixels: availableBytes / 2 (uint16 = 2 bytes per pixel)
	totalPixels := availableBytes / 2

	// Pixels per frame
	pixelsPerFrame := totalPixels / int64(numImages)

	// Dimension: sqrt(pixelsPerFrame)
	dimension := int(math.Sqrt(float64(pixelsPerFrame)))

	// Round DOWN to multiple of 256 (or 128 if < 256) to ensure we don't exceed size
	if dimension >= 256 {
		width = (dimension / 256) * 256
	} else if dimension >= 128 {
		width = 128
	} else {
		width = 128 // Minimum
	}

	height = width

	// Ensure minimum dimensions
	if width < 128 {
		width = 128
		height = 128
	}

	return width, height, nil
}

// GenerateDICOMSeries generates a complete DICOM series with multiple studies
func GenerateDICOMSeries(opts GeneratorOptions) ([]GeneratedFile, error) {
	// Validate options
	if opts.NumImages <= 0 {
		return nil, fmt.Errorf("number of images must be > 0, got %d", opts.NumImages)
	}

	// When using predefined patients, infer counts from the structure
	if len(opts.PredefinedPatients) > 0 {
		opts.NumPatients = len(opts.PredefinedPatients)
		opts.NumStudies = 0
		for _, p := range opts.PredefinedPatients {
			opts.NumStudies += len(p.Studies)
		}
	}

	if opts.NumStudies <= 0 {
		return nil, fmt.Errorf("number of studies must be > 0, got %d", opts.NumStudies)
	}
	// Default to 1 patient if not specified
	if opts.NumPatients <= 0 {
		opts.NumPatients = 1
	}
	if opts.NumPatients > opts.NumStudies {
		return nil, fmt.Errorf("number of patients (%d) cannot exceed number of studies (%d)", opts.NumPatients, opts.NumStudies)
	}

	// Parse total size
	totalBytes, err := util.ParseSize(opts.TotalSize)
	if err != nil {
		return nil, fmt.Errorf("invalid size: %w", err)
	}

	// Calculate dimensions
	width, height, err := CalculateDimensions(totalBytes, opts.NumImages)
	if err != nil {
		return nil, fmt.Errorf("calculate dimensions: %w", err)
	}

	if !opts.Quiet {
		fmt.Printf("Resolution: %dx%d pixels per image\n", width, height)
	}

	// Create output directory
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}

	// Set seed for reproducibility
	var seed int64
	if opts.Seed != 0 {
		seed = opts.Seed
		if !opts.Quiet {
			fmt.Printf("Using seed: %d\n", seed)
		}
	} else {
		// Generate deterministic seed from output directory name
		h := fnv.New64a()
		_, _ = h.Write([]byte(opts.OutputDir)) // hash.Write never returns an error
		seed = int64(h.Sum64())
		if !opts.Quiet {
			fmt.Printf("Auto-generated seed from '%s': %d\n", opts.OutputDir, seed)
			fmt.Println("  (same directory = same patient/study IDs)")
		}
	}

	// Create RNG for patient name generation
	rng := randv2.New(randv2.NewPCG(uint64(seed), uint64(seed)))

	// Create edge case applicator if enabled
	var edgeCaseApplicator *edgecases.Applicator
	if opts.EdgeCaseConfig.IsEnabled() {
		edgeCaseApplicator = edgecases.NewApplicator(opts.EdgeCaseConfig, rng)
	}

	// Create corruption applicator if enabled
	var corruptionApplicator *corruption.Applicator
	if opts.CorruptionConfig.IsEnabled() {
		corruptionApplicator = corruption.NewApplicator(opts.CorruptionConfig, rng)
	}

	// Generate or use predefined patients
	numPatients := opts.NumPatients
	if len(opts.PredefinedPatients) > 0 {
		numPatients = len(opts.PredefinedPatients)
	}
	patients := make([]patientInfo, numPatients)

	if len(opts.PredefinedPatients) > 0 {
		// Use predefined patient data from config file
		for i, p := range opts.PredefinedPatients {
			patients[i] = patientInfo{
				ID:        p.ID,
				Name:      p.Name,
				Sex:       p.Sex,
				BirthDate: p.BirthDate,
			}
			// Generate missing values
			if patients[i].Sex == "" {
				patients[i].Sex = []string{"M", "F"}[rng.IntN(2)]
			}
			if patients[i].BirthDate == "" {
				patients[i].BirthDate = fmt.Sprintf("%04d%02d%02d",
					rng.IntN(51)+1950, rng.IntN(12)+1, rng.IntN(28)+1)
			}
			if patients[i].ID == "" {
				patients[i].ID = fmt.Sprintf("PID%06d", rng.IntN(900000)+100000)
			}
			if patients[i].Name == "" {
				patients[i].Name = util.GeneratePatientName(patients[i].Sex, rng)
			}
		}
	} else {
		// Generate random patients
		for i := 0; i < numPatients; i++ {
			generatedSex := []string{"M", "F"}[rng.IntN(2)]
			generatedBirthDate := fmt.Sprintf("%04d%02d%02d",
				rng.IntN(51)+1950, // 1950-2000
				rng.IntN(12)+1,    // 1-12
				rng.IntN(28)+1)    // 1-28
			generatedID := fmt.Sprintf("PID%06d", rng.IntN(900000)+100000)
			generatedName := util.GeneratePatientName(generatedSex, rng)

			// Apply edge cases if enabled and dice roll succeeds
			if edgeCaseApplicator != nil && edgeCaseApplicator.ShouldApply() {
				generatedName = edgeCaseApplicator.ApplyToPatientName(generatedSex, generatedName)
				generatedID = edgeCaseApplicator.ApplyToPatientID(generatedID)
				generatedBirthDate = edgeCaseApplicator.ApplyToBirthDate(generatedBirthDate)
			}

			// Apply custom tags - patient-level custom tags apply to all patients
			patients[i] = patientInfo{
				ID:        getTagValue(opts.CustomTags, "PatientID", generatedID),
				Sex:       getTagValue(opts.CustomTags, "PatientSex", generatedSex),
				BirthDate: getTagValue(opts.CustomTags, "PatientBirthDate", generatedBirthDate),
				Name:      getTagValue(opts.CustomTags, "PatientName", generatedName),
			}
		}
	}

	// Generate institution info (shared or varied per study)
	var defaultInstitution util.Institution
	if !opts.VariedMetadata {
		if opts.Institution != "" {
			defaultInstitution = util.Institution{
				Name:       opts.Institution,
				Address:    "",
				Department: opts.Department,
			}
			if defaultInstitution.Department == "" {
				defaultInstitution.Department = util.Departments[rng.IntN(len(util.Departments))]
			}
		} else {
			defaultInstitution = util.GenerateInstitution(rng)
			if opts.Department != "" {
				defaultInstitution.Department = opts.Department
			}
		}
	}

	// Get modality generator
	modalityGen := modalities.GetGenerator(opts.Modality)
	modalityStr := string(modalityGen.Modality())

	// Generate body part (if fixed)
	bodyPart := opts.BodyPart
	if bodyPart == "" {
		bodyPart = util.GenerateBodyPart(modalityStr, rng)
	}

	// Generate default study-level values (used when --varied-metadata is false)
	// These are generated once and reused across all studies
	var defaultReferringPhysician, defaultPerformingPhysician, defaultOperatorName, defaultStationName string
	var defaultAccessionNumber string
	if !opts.VariedMetadata {
		defaultReferringPhysician = util.GeneratePhysicianName(rng)
		defaultPerformingPhysician = util.GeneratePhysicianName(rng)
		defaultOperatorName = util.GeneratePhysicianName(rng)
		defaultStationName = util.GenerateStationName(modalityStr, bodyPart, rng)
		defaultAccessionNumber = fmt.Sprintf("ACC%08d", rng.IntN(90000000)+10000000)
	}

	// Build patient-to-study assignment
	type studyMapping struct {
		patientIdx int
		studyIdx   int // index within patient's studies (for predefined)
	}

	var patientForStudy []studyMapping
	var numStudies int

	if len(opts.PredefinedPatients) > 0 {
		// Build mapping from predefined patients' studies
		for patientIdx, p := range opts.PredefinedPatients {
			for studyIdx := range p.Studies {
				patientForStudy = append(patientForStudy, studyMapping{
					patientIdx: patientIdx,
					studyIdx:   studyIdx,
				})
			}
		}
		numStudies = len(patientForStudy)
	} else {
		// Calculate studies per patient (distribute evenly)
		numStudies = opts.NumStudies
		studiesPerPatient := numStudies / numPatients
		remainingStudies := numStudies % numPatients

		patientForStudy = make([]studyMapping, numStudies)
		studyIdx := 0
		for patientIdx := 0; patientIdx < numPatients; patientIdx++ {
			numStudiesForThisPatient := studiesPerPatient
			if patientIdx < remainingStudies {
				numStudiesForThisPatient++
			}
			for s := 0; s < numStudiesForThisPatient; s++ {
				patientForStudy[studyIdx] = studyMapping{patientIdx: patientIdx, studyIdx: s}
				studyIdx++
			}
		}
	}

	if !opts.Quiet {
		fmt.Printf("Generating %d DICOM files...\n", opts.NumImages)
		fmt.Printf("Number of patients: %d\n", numPatients)
		// Count studies per patient from the mapping
		studyCountPerPatient := make(map[int]int)
		for _, m := range patientForStudy {
			studyCountPerPatient[m.patientIdx]++
		}
		for i, p := range patients {
			fmt.Printf("  Patient %d: %s (ID: %s, DOB: %s, Sex: %s) - %d studies\n",
				i+1, p.Name, p.ID, p.BirthDate, p.Sex, studyCountPerPatient[i])
		}
		fmt.Printf("Number of studies: %d\n", numStudies)
	}

	// Determine series per study range (default to 1 series if not specified)
	seriesPerStudy := opts.SeriesPerStudy
	if seriesPerStudy.Max == 0 {
		seriesPerStudy = util.SeriesRange{Min: 1, Max: 1}
	}
	if !opts.Quiet && seriesPerStudy.IsMultiSeries() {
		fmt.Printf("Series per study: %s\n", seriesPerStudy.String())
	}

	// Calculate images per study
	imagesPerStudy := opts.NumImages / opts.NumStudies
	remainingImages := opts.NumImages % opts.NumStudies

	// Pre-allocate task slice
	tasks := make([]imageTask, 0, opts.NumImages)
	globalImageIndex := 1

	// Get available scanners for this modality
	scanners := modalityGen.Scanners()
	pixelConfig := modalityGen.PixelConfig()

	// Phase 1: Build all tasks sequentially (maintains determinism)
	for studyNum := 1; studyNum <= opts.NumStudies; studyNum++ {
		// Get patient and study mapping for this study
		mapping := patientForStudy[studyNum-1]
		patient := patients[mapping.patientIdx]

		// Get predefined study data if available
		var predefinedStudy *PredefinedStudy
		if len(opts.PredefinedPatients) > 0 {
			predefinedStudy = &opts.PredefinedPatients[mapping.patientIdx].Studies[mapping.studyIdx]
		}

		// Generate deterministic UIDs for this study
		studyUID := util.GenerateDeterministicUID(fmt.Sprintf("%s_study_%d", opts.OutputDir, studyNum))
		// Frame of reference UID shared across all series in this study
		frameOfReferenceUID := util.GenerateDeterministicUID(fmt.Sprintf("%s_study_%d_frame", opts.OutputDir, studyNum))

		// Generate study-specific info
		studyID := fmt.Sprintf("STD%04d", rng.IntN(9000)+1000)
		var studyDescription string
		if predefinedStudy != nil && predefinedStudy.Description != "" {
			studyDescription = predefinedStudy.Description
		} else if len(opts.StudyDescriptions) > 0 && studyNum-1 < len(opts.StudyDescriptions) {
			// Use custom study description if provided
			studyDescription = opts.StudyDescriptions[studyNum-1]
		} else {
			// Auto-generate study description
			baseDescription := fmt.Sprintf("%s %s", bodyPart, modalityStr) // e.g., "HEAD CT" or "BRAIN MR"
			if opts.NumStudies > 1 {
				studyDescription = fmt.Sprintf("%s - Study %d", baseDescription, studyNum)
			} else {
				studyDescription = baseDescription
			}
			// Allow custom tag override for auto-generated descriptions
			studyDescription = getTagValue(opts.CustomTags, "StudyDescription", studyDescription)
		}

		// Generate study date and time
		studyDate := fmt.Sprintf("%04d%02d%02d",
			rng.IntN(5)+2020, // 2020-2024
			rng.IntN(12)+1,   // 1-12
			rng.IntN(28)+1)   // 1-28
		if predefinedStudy != nil && predefinedStudy.Date != "" {
			studyDate = predefinedStudy.Date
		}
		studyTime := fmt.Sprintf("%02d%02d%02d",
			rng.IntN(24),  // 0-23 hours
			rng.IntN(60),  // 0-59 minutes
			rng.IntN(60))  // 0-59 seconds

		// Select scanner for this study
		scanner := scanners[rng.IntN(len(scanners))]

		// Calculate images for this study
		numImagesThisStudy := imagesPerStudy
		if studyNum <= remainingImages {
			numImagesThisStudy++
		}

		// Categorization metadata for this study
		var studyInstitution util.Institution
		if predefinedStudy != nil && predefinedStudy.Institution != "" {
			studyInstitution = util.Institution{
				Name:       predefinedStudy.Institution,
				Department: predefinedStudy.Department,
			}
		} else if opts.VariedMetadata {
			studyInstitution = util.GenerateInstitution(rng)
		} else {
			studyInstitution = defaultInstitution
		}

		// Use predefined body part if available
		studyBodyPart := bodyPart
		if predefinedStudy != nil && predefinedStudy.BodyPart != "" {
			studyBodyPart = predefinedStudy.BodyPart
		}

		// Generate or use defaults for study-level tags
		var referringPhysician, performingPhysician, operatorName, stationName, accessionNumber string
		if predefinedStudy != nil && predefinedStudy.ReferringPhysician != "" {
			referringPhysician = predefinedStudy.ReferringPhysician
			performingPhysician = util.GeneratePhysicianName(rng)
			operatorName = util.GeneratePhysicianName(rng)
			stationName = util.GenerateStationName(modalityStr, studyBodyPart, rng)
			accessionNumber = predefinedStudy.AccessionNumber
			if accessionNumber == "" {
				accessionNumber = fmt.Sprintf("ACC%08d", rng.IntN(90000000)+10000000)
			}
		} else if opts.VariedMetadata {
			// Generate new values per study when varied
			referringPhysician = util.GeneratePhysicianName(rng)
			performingPhysician = util.GeneratePhysicianName(rng)
			operatorName = util.GeneratePhysicianName(rng)
			stationName = util.GenerateStationName(modalityStr, studyBodyPart, rng)
			accessionNumber = fmt.Sprintf("ACC%08d", rng.IntN(90000000)+10000000)
		} else {
			// Use defaults (same across all studies)
			referringPhysician = defaultReferringPhysician
			performingPhysician = defaultPerformingPhysician
			operatorName = defaultOperatorName
			stationName = defaultStationName
			accessionNumber = defaultAccessionNumber
		}

		// Apply custom tag overrides for study-level tags
		institutionName := getTagValue(opts.CustomTags, "InstitutionName", studyInstitution.Name)
		institutionalDepartmentName := getTagValue(opts.CustomTags, "InstitutionalDepartmentName", studyInstitution.Department)
		referringPhysician = getTagValue(opts.CustomTags, "ReferringPhysicianName", referringPhysician)
		performingPhysician = getTagValue(opts.CustomTags, "PerformingPhysicianName", performingPhysician)
		operatorName = getTagValue(opts.CustomTags, "OperatorsName", operatorName)
		stationName = getTagValue(opts.CustomTags, "StationName", stationName)
		accessionNumber = getTagValue(opts.CustomTags, "AccessionNumber", accessionNumber)

		// Use predefined priority or default
		studyPriority := opts.Priority.String()
		if predefinedStudy != nil && predefinedStudy.Priority != "" {
			studyPriority = predefinedStudy.Priority
		}
		requestedProcedurePriority := getTagValue(opts.CustomTags, "RequestedProcedurePriority", studyPriority)

		// Generate series-level tags with custom overrides
		protocolName := util.GenerateProtocolName(modalityStr, studyBodyPart, rng)
		clinicalIndication := util.GenerateClinicalIndication(modalityStr, studyBodyPart, rng)

		// Apply custom tag overrides for series-level tags
		protocolName = getTagValue(opts.CustomTags, "ProtocolName", protocolName)
		bodyPartExamined := getTagValue(opts.CustomTags, "BodyPartExamined", studyBodyPart)
		requestedProcedureDescription := getTagValue(opts.CustomTags, "RequestedProcedureDescription", clinicalIndication)

		// Determine number of series for this study
		var numSeriesThisStudy int
		if predefinedStudy != nil && len(predefinedStudy.Series) > 0 {
			numSeriesThisStudy = len(predefinedStudy.Series)
		} else {
			numSeriesThisStudy = seriesPerStudy.GetSeriesCount(rng)
		}

		// Get series templates for this modality
		seriesTemplates := modalities.GetSeriesTemplates(opts.Modality, studyBodyPart, numSeriesThisStudy, rng)
		if predefinedStudy == nil || len(predefinedStudy.Series) == 0 {
			numSeriesThisStudy = len(seriesTemplates) // May be limited by available templates
		}

		// Ensure at least 1 series
		if numSeriesThisStudy < 1 {
			numSeriesThisStudy = 1
		}

		// Generate base modality-specific parameters for this study (shared across all series)
		baseSeriesParams := modalityGen.GenerateSeriesParams(scanner, rng)

		if !opts.Quiet {
			fmt.Printf("\nStudy %d/%d: %d images in %d series (Patient: %s)\n", studyNum, opts.NumStudies, numImagesThisStudy, numSeriesThisStudy, patient.Name)
			fmt.Printf("  StudyID: %s, Description: %s\n", studyID, studyDescription)
			fmt.Printf("  Modality: %s, Scanner: %s %s\n", modalityStr, scanner.Manufacturer, scanner.Model)
			fmt.Printf("  Resolution: PixelSpacing=%.2fmm, SliceThickness=%.2fmm\n",
				baseSeriesParams.PixelSpacing, baseSeriesParams.SliceThickness)
		}

		// Distribute images across series
		imagesPerSeries := numImagesThisStudy / numSeriesThisStudy
		remainingSeriesImages := numImagesThisStudy % numSeriesThisStudy

		instanceInStudy := 1

		// Generate images for each series
		for seriesNum := 1; seriesNum <= numSeriesThisStudy; seriesNum++ {
			// Generate deterministic series UID
			seriesUID := util.GenerateDeterministicUID(fmt.Sprintf("%s_study_%d_series_%d", opts.OutputDir, studyNum, seriesNum))

			// Get predefined series if available
			var predefinedSeries *PredefinedSeries
			if predefinedStudy != nil && seriesNum <= len(predefinedStudy.Series) {
				predefinedSeries = &predefinedStudy.Series[seriesNum-1]
			}

			// Get series template (if available)
			var seriesTemplate modalities.SeriesTemplate
			var predefinedProtocol string
			if predefinedSeries != nil {
				// Build template from predefined data
				seriesTemplate = modalities.SeriesTemplate{
					SeriesDescription: predefinedSeries.Description,
				}
				predefinedProtocol = predefinedSeries.Protocol
				// Parse orientation if provided
				switch predefinedSeries.Orientation {
				case "Sagittal", "sagittal", "SAG":
					seriesTemplate.Orientation = modalities.OrientationSagittal
				case "Coronal", "coronal", "COR":
					seriesTemplate.Orientation = modalities.OrientationCoronal
				default:
					seriesTemplate.Orientation = modalities.OrientationAxial
				}
			} else if seriesNum <= len(seriesTemplates) {
				seriesTemplate = seriesTemplates[seriesNum-1]
			} else {
				// Fallback template
				seriesTemplate = modalities.SeriesTemplate{
					SeriesDescription: fmt.Sprintf("Series %d", seriesNum),
					Orientation:       modalities.OrientationAxial,
				}
			}

			// Copy base parameters and apply series-specific overrides
			seriesParams := baseSeriesParams

			// Apply series template window settings if specified
			if seriesTemplate.WindowCenter != 0 {
				seriesParams.WindowCenter = seriesTemplate.WindowCenter
			}
			if seriesTemplate.WindowWidth != 0 {
				seriesParams.WindowWidth = seriesTemplate.WindowWidth
			}

			// Calculate images for this series
			var numImagesThisSeries int
			if predefinedSeries != nil && predefinedSeries.ImageCount > 0 {
				numImagesThisSeries = predefinedSeries.ImageCount
			} else {
				numImagesThisSeries = imagesPerSeries
				if seriesNum <= remainingSeriesImages {
					numImagesThisSeries++
				}
			}

			// Generate series description
			generatedSeriesDescription := seriesTemplate.SeriesDescription
			if generatedSeriesDescription == "" {
				generatedSeriesDescription = fmt.Sprintf("Series %d - %s", seriesNum, modalityStr)
			}
			seriesDescription := getTagValue(opts.CustomTags, "SeriesDescription", generatedSeriesDescription)

			// Use series-specific protocol if available
			seriesProtocolName := protocolName
			if predefinedProtocol != "" {
				seriesProtocolName = predefinedProtocol
			}

			// Get image orientation from template
			imageOrientationValues := seriesTemplate.ImageOrientationPatient()
			imageOrientationPatient := make([]string, 6)
			for i, v := range imageOrientationValues {
				imageOrientationPatient[i] = fmt.Sprintf("%.6f", v)
			}

			if !opts.Quiet {
				fmt.Printf("  Series %d: %s (%d images, %s)\n", seriesNum, seriesDescription, numImagesThisSeries, seriesTemplate.Orientation)
			}

			// Build tasks for each image in this series
			for instanceInSeries := 1; instanceInSeries <= numImagesThisSeries; instanceInSeries++ {
				sopInstanceUID := util.GenerateDeterministicUID(
					fmt.Sprintf("%s_study_%d_series_%d_instance_%d", opts.OutputDir, studyNum, seriesNum, instanceInSeries))

				sliceIndex := float64(instanceInSeries - 1)
				imagePositionX := -100.0
				imagePositionY := -100.0
				imagePositionZ := -100.0 + (sliceIndex * seriesParams.SpacingBetweenSlices)
				imagePositionPatient := []string{
					fmt.Sprintf("%.6f", imagePositionX),
					fmt.Sprintf("%.6f", imagePositionY),
					fmt.Sprintf("%.6f", imagePositionZ),
				}
				sliceLocation := imagePositionZ

				// Build metadata (without pixel data)
				metadata := []*dicom.Element{
					mustNewElement(tag.TransferSyntaxUID, []string{"1.2.840.10008.1.2.1"}),
					mustNewElement(tag.PatientName, []string{patient.Name}),
					mustNewElement(tag.PatientID, []string{patient.ID}),
					mustNewElement(tag.PatientBirthDate, []string{patient.BirthDate}),
					mustNewElement(tag.PatientSex, []string{patient.Sex}),
					mustNewElement(tag.StudyInstanceUID, []string{studyUID}),
					mustNewElement(tag.StudyID, []string{studyID}),
					mustNewElement(tag.StudyDate, []string{studyDate}),
					mustNewElement(tag.StudyTime, []string{studyTime}),
					mustNewElement(tag.StudyDescription, []string{studyDescription}),
					mustNewElement(tag.SeriesInstanceUID, []string{seriesUID}),
					mustNewElement(tag.SeriesNumber, []string{fmt.Sprintf("%d", seriesNum)}),
					mustNewElement(tag.SeriesDescription, []string{seriesDescription}),
					mustNewElement(tag.Modality, []string{modalityStr}),
					mustNewElement(tag.SOPInstanceUID, []string{sopInstanceUID}),
					mustNewElement(tag.SOPClassUID, []string{modalityGen.SOPClassUID()}),
					mustNewElement(tag.InstanceNumber, []string{fmt.Sprintf("%d", instanceInSeries)}),
					mustNewElement(tag.PixelSpacing, []string{
						fmt.Sprintf("%.6f", seriesParams.PixelSpacing),
						fmt.Sprintf("%.6f", seriesParams.PixelSpacing),
					}),
					mustNewElement(tag.SliceThickness, []string{fmt.Sprintf("%.6f", seriesParams.SliceThickness)}),
					mustNewElement(tag.SpacingBetweenSlices, []string{fmt.Sprintf("%.6f", seriesParams.SpacingBetweenSlices)}),
					mustNewElement(tag.Manufacturer, []string{scanner.Manufacturer}),
					mustNewElement(tag.ManufacturerModelName, []string{scanner.Model}),
					mustNewElement(tag.WindowCenter, []string{fmt.Sprintf("%.1f", seriesParams.WindowCenter)}),
					mustNewElement(tag.WindowWidth, []string{fmt.Sprintf("%.1f", seriesParams.WindowWidth)}),
					mustNewElement(tag.ImagePositionPatient, imagePositionPatient),
					mustNewElement(tag.ImageOrientationPatient, imageOrientationPatient),
					mustNewElement(tag.SliceLocation, []string{fmt.Sprintf("%.6f", sliceLocation)}),
					mustNewElement(tag.FrameOfReferenceUID, []string{frameOfReferenceUID}),
					mustNewElement(tag.Rows, []int{height}),
					mustNewElement(tag.Columns, []int{width}),
					mustNewElement(tag.BitsAllocated, []int{int(pixelConfig.BitsAllocated)}),
					mustNewElement(tag.BitsStored, []int{int(pixelConfig.BitsStored)}),
					mustNewElement(tag.HighBit, []int{int(pixelConfig.HighBit)}),
					mustNewElement(tag.PixelRepresentation, []int{int(pixelConfig.PixelRepresentation)}),
					mustNewElement(tag.SamplesPerPixel, []int{1}),
					mustNewElement(tag.PhotometricInterpretation, []string{"MONOCHROME2"}),
					// Categorization tags (with custom tag overrides applied)
					mustNewElement(tag.InstitutionName, []string{institutionName}),
					mustNewElement(tag.InstitutionalDepartmentName, []string{institutionalDepartmentName}),
					mustNewElement(tag.StationName, []string{stationName}),
					mustNewElement(tag.ReferringPhysicianName, []string{referringPhysician}),
					mustNewElement(tag.PerformingPhysicianName, []string{performingPhysician}),
					mustNewElement(tag.OperatorsName, []string{operatorName}),
					mustNewElement(tag.BodyPartExamined, []string{bodyPartExamined}),
					mustNewElement(tag.ProtocolName, []string{seriesProtocolName}),
					mustNewElement(tag.RequestedProcedureDescription, []string{requestedProcedureDescription}),
					mustNewElement(tag.RequestedProcedurePriority, []string{requestedProcedurePriority}),
					mustNewElement(tag.AccessionNumber, []string{accessionNumber}),
				}

				// Add contrast agent info if this series uses contrast
				if seriesTemplate.HasContrast && seriesTemplate.ContrastAgent != "" {
					metadata = append(metadata, mustNewElement(tag.ContrastBolusAgent, []string{seriesTemplate.ContrastAgent}))
				}

				// Add sequence name for MR
				if seriesTemplate.SequenceName != "" {
					metadata = append(metadata, mustNewElement(tag.SequenceName, []string{seriesTemplate.SequenceName}))
				}

				// Add modality-specific elements
				ds := &dicom.Dataset{Elements: metadata}
				if err := modalityGen.AppendModalityElements(ds, seriesParams); err != nil {
					return nil, fmt.Errorf("add modality elements for study %d, series %d, instance %d: %w", studyNum, seriesNum, instanceInSeries, err)
				}
				metadata = ds.Elements

				// Add corruption elements if enabled
				var taskWriteOpts []dicom.WriteOption
				var taskHasMalformedLengths bool
				if corruptionApplicator != nil {
					corruptionElements := corruptionApplicator.GenerateCorruptionElements()
					metadata = append(metadata, corruptionElements...)

					// Sort metadata by (Group, Element) so private tags (e.g., 0x0009)
					// are placed before standard tags they might precede
					sort.Slice(metadata, func(i, j int) bool {
						if metadata[i].Tag.Group != metadata[j].Tag.Group {
							return metadata[i].Tag.Group < metadata[j].Tag.Group
						}
						return metadata[i].Tag.Element < metadata[j].Tag.Element
					})

					taskWriteOpts = []dicom.WriteOption{dicom.SkipVRVerification(), dicom.SkipValueTypeVerification()}
					taskHasMalformedLengths = corruptionApplicator.HasMalformedLengths()
				}

				// Generate deterministic pixel seed for this specific image
				pixelSeedHash := fnv.New64a()
				_, _ = pixelSeedHash.Write([]byte(fmt.Sprintf("%d_pixel_%d", seed, globalImageIndex)))
				pixelSeed := pixelSeedHash.Sum64()

				filename := fmt.Sprintf("IMG%04d.dcm", globalImageIndex)
				filePath := filepath.Join(opts.OutputDir, filename)

				tasks = append(tasks, imageTask{
					globalIndex:         globalImageIndex,
					instanceInStudy:     instanceInStudy,
					instanceInSeries:    instanceInSeries,
					seriesNumber:        seriesNum,
					width:               width,
					height:              height,
					filePath:            filePath,
					textOverlay:         fmt.Sprintf("File %d/%d", globalImageIndex, opts.NumImages),
					pixelSeed:           pixelSeed,
					metadata:            metadata,
					pixelConfig:         pixelConfig,
					writeOpts:           taskWriteOpts,
					hasMalformedLengths: taskHasMalformedLengths,
					studyUID:            studyUID,
					seriesUID:           seriesUID,
					sopInstanceUID:      sopInstanceUID,
					patientID:           patient.ID,
					studyID:             studyID,
				})

				globalImageIndex++
				instanceInStudy++
			}
		}
	}

	// Phase 2: Process tasks in parallel
	numWorkers := opts.Workers
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}
	// Don't use more workers than tasks
	if numWorkers > len(tasks) {
		numWorkers = len(tasks)
	}

	if !opts.Quiet {
		fmt.Printf("\nGenerating images with %d parallel workers...\n", numWorkers)
	}

	// Create channels for work distribution and results
	taskChan := make(chan imageTask, len(tasks))
	resultChan := make(chan struct {
		index int
		err   error
	}, len(tasks))

	// Start workers
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChan {
				err := generateImageFromTask(task)
				resultChan <- struct {
					index int
					err   error
				}{task.globalIndex, err}
			}
		}()
	}

	// Send all tasks to workers
	for _, task := range tasks {
		taskChan <- task
	}
	close(taskChan)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results and track progress
	completed := 0
	var firstErr error
	for result := range resultChan {
		if result.err != nil && firstErr == nil {
			firstErr = fmt.Errorf("generate image %d: %w", result.index, result.err)
		}
		completed++
		// Call progress callback if provided
		if opts.ProgressCallback != nil {
			opts.ProgressCallback(completed, len(tasks))
		}
		if !opts.Quiet && (completed%10 == 0 || completed == len(tasks)) {
			progress := float64(completed) / float64(len(tasks)) * 100
			fmt.Printf("  Progress: %d/%d (%.0f%%)\n", completed, len(tasks), progress)
		}
	}

	if firstErr != nil {
		return nil, firstErr
	}

	// Build result slice (in order)
	generatedFiles := make([]GeneratedFile, len(tasks))
	for i, task := range tasks {
		generatedFiles[i] = GeneratedFile{
			Path:            task.filePath,
			StudyUID:        task.studyUID,
			SeriesUID:       task.seriesUID,
			SOPInstanceUID:  task.sopInstanceUID,
			PatientID:       task.patientID,
			StudyID:         task.studyID,
			SeriesNumber:    task.seriesNumber,
			InstanceNumber:  task.instanceInSeries,
			InstanceInStudy: task.instanceInStudy,
		}
	}

	if !opts.Quiet {
		fmt.Printf("\n✓ %d DICOM files created in: %s/\n", opts.NumImages, opts.OutputDir)
	}

	return generatedFiles, nil
}
