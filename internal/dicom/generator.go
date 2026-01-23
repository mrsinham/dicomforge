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

	"github.com/mrsinham/dicomforge/internal/dicom/edgecases"
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
func writeDatasetToFile(filename string, ds dicom.Dataset) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return dicom.Write(f, ds)
}

// drawTextOnFrame draws large text overlay on a uint16 frame
func drawTextOnFrame(nativeFrame *frame.NativeFrame[uint16], width, height int, text string) {
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

// GeneratorOptions contains all parameters needed to generate a DICOM series
type GeneratorOptions struct {
	NumImages   int
	TotalSize   string
	OutputDir   string
	Seed        int64
	NumStudies  int
	NumPatients int // Number of patients (studies are distributed among patients)
	Workers     int // Number of parallel workers (0 = auto-detect based on CPU cores)

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
	globalIndex     int
	instanceInStudy int
	width           int
	height          int
	filePath        string
	textOverlay     string
	pixelSeed       uint64 // Deterministic seed for this image's pixel generation
	metadata        []*dicom.Element
	// Result info
	studyUID       string
	seriesUID      string
	sopInstanceUID string
	patientID      string
	studyID        string
}

// GeneratedFile contains information about a generated DICOM file
type GeneratedFile struct {
	Path           string
	StudyUID       string
	SeriesUID      string
	SOPInstanceUID string
	PatientID      string
	StudyID        string
	SeriesNumber   int
	InstanceNumber int
}

// generateImageFromTask generates a single DICOM image from a pre-computed task
func generateImageFromTask(task imageTask) error {
	width, height := task.width, task.height
	pixelsPerFrame := width * height

	// Create frame
	nativeFrame := frame.NewNativeFrame[uint16](16, height, width, pixelsPerFrame, 1)

	// Create deterministic RNG for this specific image
	rng := randv2.New(randv2.NewPCG(task.pixelSeed, task.pixelSeed))

	// Fill with synthetic brain-like pattern with noise everywhere
	centerX, centerY := float64(width)/2, float64(height)/2
	maxDist := math.Sqrt(centerX*centerX + centerY*centerY)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			dist := math.Sqrt(dx*dx + dy*dy)

			baseIntensity := (1.0 - (dist / maxDist)) * 12000.0

			largeNoise := (rng.Float64() - 0.5) * 12000.0
			mediumNoise := (rng.Float64() - 0.5) * 6000.0
			fineNoise := (rng.Float64() - 0.5) * 3000.0

			totalNoise := largeNoise + mediumNoise + fineNoise
			intensity := baseIntensity + totalNoise

			pixelValue := uint16(math.Max(0, math.Min(65535, intensity)))
			nativeFrame.RawData[y*width+x] = pixelValue
		}
	}

	// Draw text overlay
	drawTextOnFrame(nativeFrame, width, height, task.textOverlay)

	// Create pixel data info
	pixelDataInfo := dicom.PixelDataInfo{
		Frames: []*frame.Frame{
			{
				Encapsulated: false,
				NativeData:   nativeFrame,
			},
		},
	}

	// Build complete metadata with pixel data
	elements := make([]*dicom.Element, len(task.metadata)+1)
	copy(elements, task.metadata)
	elements[len(task.metadata)] = mustNewElement(tag.PixelData, pixelDataInfo)

	// Write DICOM file
	return writeDatasetToFile(task.filePath, dicom.Dataset{Elements: elements})
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

	fmt.Printf("Resolution: %dx%d pixels per image\n", width, height)

	// Create output directory
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}

	// Set seed for reproducibility
	var seed int64
	if opts.Seed != 0 {
		seed = opts.Seed
		fmt.Printf("Using seed: %d\n", seed)
	} else {
		// Generate deterministic seed from output directory name
		h := fnv.New64a()
		_, _ = h.Write([]byte(opts.OutputDir)) // hash.Write never returns an error
		seed = int64(h.Sum64())
		fmt.Printf("Auto-generated seed from '%s': %d\n", opts.OutputDir, seed)
		fmt.Println("  (same directory = same patient/study IDs)")
	}

	// Create RNG for patient name generation
	rng := randv2.New(randv2.NewPCG(uint64(seed), uint64(seed)))

	// Create edge case applicator if enabled
	var edgeCaseApplicator *edgecases.Applicator
	if opts.EdgeCaseConfig.IsEnabled() {
		edgeCaseApplicator = edgecases.NewApplicator(opts.EdgeCaseConfig, rng)
	}

	// Generate all patients
	patients := make([]patientInfo, opts.NumPatients)
	for i := 0; i < opts.NumPatients; i++ {
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

	// Generate body part (if fixed)
	bodyPart := opts.BodyPart
	if bodyPart == "" {
		bodyPart = util.GenerateBodyPart("MR", rng)
	}

	// Generate default study-level values (used when --varied-metadata is false)
	// These are generated once and reused across all studies
	var defaultReferringPhysician, defaultPerformingPhysician, defaultOperatorName, defaultStationName string
	var defaultAccessionNumber string
	if !opts.VariedMetadata {
		defaultReferringPhysician = util.GeneratePhysicianName(rng)
		defaultPerformingPhysician = util.GeneratePhysicianName(rng)
		defaultOperatorName = util.GeneratePhysicianName(rng)
		defaultStationName = util.GenerateStationName("MR", bodyPart, rng)
		defaultAccessionNumber = fmt.Sprintf("ACC%08d", rng.IntN(90000000)+10000000)
	}

	// Calculate studies per patient (distribute evenly)
	studiesPerPatient := opts.NumStudies / opts.NumPatients
	remainingStudies := opts.NumStudies % opts.NumPatients

	// Build patient-to-study assignment
	// Each patient gets at least studiesPerPatient studies
	// First remainingStudies patients get one extra
	patientForStudy := make([]int, opts.NumStudies)
	studyIdx := 0
	for patientIdx := 0; patientIdx < opts.NumPatients; patientIdx++ {
		numStudiesForThisPatient := studiesPerPatient
		if patientIdx < remainingStudies {
			numStudiesForThisPatient++
		}
		for s := 0; s < numStudiesForThisPatient; s++ {
			patientForStudy[studyIdx] = patientIdx
			studyIdx++
		}
	}

	fmt.Printf("Generating %d DICOM files...\n", opts.NumImages)
	fmt.Printf("Number of patients: %d\n", opts.NumPatients)
	for i, p := range patients {
		studyCount := studiesPerPatient
		if i < remainingStudies {
			studyCount++
		}
		fmt.Printf("  Patient %d: %s (ID: %s, DOB: %s, Sex: %s) - %d studies\n",
			i+1, p.Name, p.ID, p.BirthDate, p.Sex, studyCount)
	}
	fmt.Printf("Number of studies: %d\n", opts.NumStudies)

	// Calculate images per study
	imagesPerStudy := opts.NumImages / opts.NumStudies
	remainingImages := opts.NumImages % opts.NumStudies

	// Pre-allocate task slice
	tasks := make([]imageTask, 0, opts.NumImages)
	globalImageIndex := 1

	// Manufacturer options
	manufacturers := []struct {
		Name          string
		Model         string
		FieldStrength float64
	}{
		{"SIEMENS", "Avanto", 1.5},
		{"SIEMENS", "Skyra", 3.0},
		{"GE MEDICAL SYSTEMS", "Signa HDxt", 1.5},
		{"GE MEDICAL SYSTEMS", "Discovery MR750", 3.0},
		{"PHILIPS", "Achieva", 1.5},
		{"PHILIPS", "Ingenia", 3.0},
	}

	// Phase 1: Build all tasks sequentially (maintains determinism)
	for studyNum := 1; studyNum <= opts.NumStudies; studyNum++ {
		// Get patient for this study
		patient := patients[patientForStudy[studyNum-1]]

		// Generate deterministic UIDs for this study
		studyUID := util.GenerateDeterministicUID(fmt.Sprintf("%s_study_%d", opts.OutputDir, studyNum))
		seriesUID := util.GenerateDeterministicUID(fmt.Sprintf("%s_study_%d_series_1", opts.OutputDir, studyNum))

		// Generate study-specific info
		studyID := fmt.Sprintf("STD%04d", rng.IntN(9000)+1000)
		var generatedStudyDescription string
		if opts.NumStudies > 1 {
			generatedStudyDescription = fmt.Sprintf("Brain MRI - Study %d", studyNum)
		} else {
			generatedStudyDescription = "Brain MRI"
		}
		studyDescription := getTagValue(opts.CustomTags, "StudyDescription", generatedStudyDescription)

		// Generate study date and time
		studyDate := fmt.Sprintf("%04d%02d%02d",
			rng.IntN(5)+2020, // 2020-2024
			rng.IntN(12)+1,   // 1-12
			rng.IntN(28)+1)   // 1-28
		studyTime := fmt.Sprintf("%02d%02d%02d",
			rng.IntN(24),  // 0-23 hours
			rng.IntN(60),  // 0-59 minutes
			rng.IntN(60))  // 0-59 seconds

		// Generate series-specific MRI parameters (same for all images in series)
		seriesPixelSpacing := rng.Float64()*1.5 + 0.5                           // 0.5-2.0 mm
		seriesSliceThickness := rng.Float64()*4.0 + 1.0                         // 1.0-5.0 mm
		seriesSpacingBetweenSlices := seriesSliceThickness + rng.Float64()*0.5  // slightly larger than thickness
		seriesEchoTime := rng.Float64()*20.0 + 10.0                             // 10-30 ms
		seriesRepetitionTime := rng.Float64()*400.0 + 400.0                     // 400-800 ms
		seriesFlipAngle := rng.Float64()*30.0 + 60.0                            // 60-90 degrees
		seriesSequenceName := []string{"T1_MPRAGE", "T1_SE", "T2_FSE", "T2_FLAIR"}[rng.IntN(4)]

		// Select MRI scanner
		mfr := manufacturers[rng.IntN(len(manufacturers))]

		// Calculate imaging frequency based on field strength (in MHz)
		imagingFrequency := mfr.FieldStrength * 42.58

		// Window settings for display
		windowCenter := 500.0 + rng.Float64()*1000.0 // 500-1500
		windowWidth := 1000.0 + rng.Float64()*1000.0 // 1000-2000

		// Calculate images for this study
		numImagesThisStudy := imagesPerStudy
		if studyNum <= remainingImages {
			numImagesThisStudy++
		}

		fmt.Printf("\nStudy %d/%d: %d images (Patient: %s)\n", studyNum, opts.NumStudies, numImagesThisStudy, patient.Name)
		fmt.Printf("  StudyID: %s, Description: %s\n", studyID, studyDescription)
		fmt.Printf("  Scanner: %s %s (%.1fT)\n", mfr.Name, mfr.Model, mfr.FieldStrength)
		fmt.Printf("  Sequence: %s, TE=%.1fms, TR=%.1fms, FA=%.1f°\n",
			seriesSequenceName, seriesEchoTime, seriesRepetitionTime, seriesFlipAngle)
		fmt.Printf("  Resolution: PixelSpacing=%.2fmm, SliceThickness=%.2fmm\n",
			seriesPixelSpacing, seriesSliceThickness)

		// Categorization metadata for this study
		var studyInstitution util.Institution
		if opts.VariedMetadata {
			studyInstitution = util.GenerateInstitution(rng)
		} else {
			studyInstitution = defaultInstitution
		}

		// Generate or use defaults for study-level tags
		var referringPhysician, performingPhysician, operatorName, stationName, accessionNumber string
		if opts.VariedMetadata {
			// Generate new values per study when varied
			referringPhysician = util.GeneratePhysicianName(rng)
			performingPhysician = util.GeneratePhysicianName(rng)
			operatorName = util.GeneratePhysicianName(rng)
			stationName = util.GenerateStationName("MR", bodyPart, rng)
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
		requestedProcedurePriority := getTagValue(opts.CustomTags, "RequestedProcedurePriority", opts.Priority.String())

		// Generate series-level tags with custom overrides
		protocolName := util.GenerateProtocolName("MR", bodyPart, rng)
		clinicalIndication := util.GenerateClinicalIndication("MR", bodyPart, rng)
		generatedSeriesDescription := fmt.Sprintf("Series 1 - %s", seriesSequenceName)

		// Apply custom tag overrides for series-level tags
		protocolName = getTagValue(opts.CustomTags, "ProtocolName", protocolName)
		bodyPartExamined := getTagValue(opts.CustomTags, "BodyPartExamined", bodyPart)
		requestedProcedureDescription := getTagValue(opts.CustomTags, "RequestedProcedureDescription", clinicalIndication)
		seriesDescription := getTagValue(opts.CustomTags, "SeriesDescription", generatedSeriesDescription)

		// Build tasks for each image in this study
		for instanceInStudy := 1; instanceInStudy <= numImagesThisStudy; instanceInStudy++ {
			sopInstanceUID := util.GenerateDeterministicUID(
				fmt.Sprintf("%s_study_%d_instance_%d", opts.OutputDir, studyNum, instanceInStudy))

			imageOrientationPatient := []string{"1", "0", "0", "0", "1", "0"}

			sliceIndex := float64(instanceInStudy - 1)
			imagePositionX := -100.0
			imagePositionY := -100.0
			imagePositionZ := -100.0 + (sliceIndex * seriesSpacingBetweenSlices)
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
				mustNewElement(tag.SeriesNumber, []string{fmt.Sprintf("%d", 1)}),
				mustNewElement(tag.SeriesDescription, []string{seriesDescription}),
				mustNewElement(tag.Modality, []string{"MR"}),
				mustNewElement(tag.SOPInstanceUID, []string{sopInstanceUID}),
				mustNewElement(tag.SOPClassUID, []string{"1.2.840.10008.5.1.4.1.1.4"}),
				mustNewElement(tag.InstanceNumber, []string{fmt.Sprintf("%d", instanceInStudy)}),
				mustNewElement(tag.PixelSpacing, []string{
					fmt.Sprintf("%.6f", seriesPixelSpacing),
					fmt.Sprintf("%.6f", seriesPixelSpacing),
				}),
				mustNewElement(tag.SliceThickness, []string{fmt.Sprintf("%.6f", seriesSliceThickness)}),
				mustNewElement(tag.SpacingBetweenSlices, []string{fmt.Sprintf("%.6f", seriesSpacingBetweenSlices)}),
				mustNewElement(tag.EchoTime, []string{fmt.Sprintf("%.6f", seriesEchoTime)}),
				mustNewElement(tag.RepetitionTime, []string{fmt.Sprintf("%.6f", seriesRepetitionTime)}),
				mustNewElement(tag.FlipAngle, []string{fmt.Sprintf("%.6f", seriesFlipAngle)}),
				mustNewElement(tag.MagneticFieldStrength, []string{fmt.Sprintf("%.1f", mfr.FieldStrength)}),
				mustNewElement(tag.ImagingFrequency, []string{fmt.Sprintf("%.6f", imagingFrequency)}),
				mustNewElement(tag.Manufacturer, []string{mfr.Name}),
				mustNewElement(tag.ManufacturerModelName, []string{mfr.Model}),
				mustNewElement(tag.SequenceName, []string{seriesSequenceName}),
				mustNewElement(tag.WindowCenter, []string{fmt.Sprintf("%.1f", windowCenter)}),
				mustNewElement(tag.WindowWidth, []string{fmt.Sprintf("%.1f", windowWidth)}),
				mustNewElement(tag.ImagePositionPatient, imagePositionPatient),
				mustNewElement(tag.ImageOrientationPatient, imageOrientationPatient),
				mustNewElement(tag.SliceLocation, []string{fmt.Sprintf("%.6f", sliceLocation)}),
				mustNewElement(tag.Rows, []int{height}),
				mustNewElement(tag.Columns, []int{width}),
				mustNewElement(tag.BitsAllocated, []int{16}),
				mustNewElement(tag.BitsStored, []int{16}),
				mustNewElement(tag.HighBit, []int{15}),
				mustNewElement(tag.PixelRepresentation, []int{0}),
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
				mustNewElement(tag.ProtocolName, []string{protocolName}),
				mustNewElement(tag.RequestedProcedureDescription, []string{requestedProcedureDescription}),
				mustNewElement(tag.RequestedProcedurePriority, []string{requestedProcedurePriority}),
				mustNewElement(tag.AccessionNumber, []string{accessionNumber}),
			}

			// Generate deterministic pixel seed for this specific image
			pixelSeedHash := fnv.New64a()
			_, _ = pixelSeedHash.Write([]byte(fmt.Sprintf("%d_pixel_%d", seed, globalImageIndex)))
			pixelSeed := pixelSeedHash.Sum64()

			filename := fmt.Sprintf("IMG%04d.dcm", globalImageIndex)
			filePath := filepath.Join(opts.OutputDir, filename)

			tasks = append(tasks, imageTask{
				globalIndex:     globalImageIndex,
				instanceInStudy: instanceInStudy,
				width:           width,
				height:          height,
				filePath:        filePath,
				textOverlay:     fmt.Sprintf("File %d/%d", globalImageIndex, opts.NumImages),
				pixelSeed:       pixelSeed,
				metadata:        metadata,
				studyUID:        studyUID,
				seriesUID:       seriesUID,
				sopInstanceUID:  sopInstanceUID,
				patientID:       patient.ID,
				studyID:         studyID,
			})

			globalImageIndex++
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

	fmt.Printf("\nGenerating images with %d parallel workers...\n", numWorkers)

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
		if completed%10 == 0 || completed == len(tasks) {
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
			Path:           task.filePath,
			StudyUID:       task.studyUID,
			SeriesUID:      task.seriesUID,
			SOPInstanceUID: task.sopInstanceUID,
			PatientID:      task.patientID,
			StudyID:        task.studyID,
			SeriesNumber:   1,
			InstanceNumber: task.instanceInStudy,
		}
	}

	fmt.Printf("\n✓ %d DICOM files created in: %s/\n", opts.NumImages, opts.OutputDir)

	return generatedFiles, nil
}
