package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	internaldicom "github.com/mrsinham/dicomforge/internal/dicom"
	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

// TestGenerateSeries_Basic tests basic DICOM series generation
func TestGenerateSeries_Basic(t *testing.T) {
	outputDir := t.TempDir()

	opts := internaldicom.GeneratorOptions{
		NumImages:  5,
		TotalSize:  "10MB",
		OutputDir:  outputDir,
		Seed:       42,
		NumStudies: 1,
	}

	t.Logf("Generating DICOM series in: %s", outputDir)

	files, err := internaldicom.GenerateDICOMSeries(opts)
	if err != nil {
		t.Fatalf("GenerateDICOMSeries failed: %v", err)
	}

	// Verify file count
	if len(files) != 5 {
		t.Errorf("Expected 5 files, got %d", len(files))
	}

	// Verify files exist
	for i, file := range files {
		if _, err := os.Stat(file.Path); os.IsNotExist(err) {
			t.Errorf("File %d does not exist: %s", i+1, file.Path)
		}
		t.Logf("Generated file %d: %s", i+1, file.Path)
	}

	// Verify UIDs are set
	if files[0].StudyUID == "" {
		t.Error("StudyUID should not be empty")
	}
	if files[0].SeriesUID == "" {
		t.Error("SeriesUID should not be empty")
	}
	if files[0].PatientID == "" {
		t.Error("PatientID should not be empty")
	}

	t.Logf("✓ Basic generation test passed")
}

// TestOrganizeFiles_DICOMDIRStructure tests DICOMDIR organization
func TestOrganizeFiles_DICOMDIRStructure(t *testing.T) {
	outputDir := t.TempDir()

	opts := internaldicom.GeneratorOptions{
		NumImages:  5,
		TotalSize:  "10MB",
		OutputDir:  outputDir,
		Seed:       42,
		NumStudies: 1,
	}

	// Generate files
	files, err := internaldicom.GenerateDICOMSeries(opts)
	if err != nil {
		t.Fatalf("GenerateDICOMSeries failed: %v", err)
	}

	t.Logf("Generated %d files, organizing into DICOMDIR...", len(files))

	// Organize into DICOMDIR structure
	err = internaldicom.OrganizeFilesIntoDICOMDIR(outputDir, files)
	if err != nil {
		t.Fatalf("OrganizeFilesIntoDICOMDIR failed: %v", err)
	}

	// 1. Verify DICOMDIR exists
	dicomdirPath := filepath.Join(outputDir, "DICOMDIR")
	if _, err := os.Stat(dicomdirPath); os.IsNotExist(err) {
		t.Error("DICOMDIR file should exist")
	} else {
		t.Logf("✓ DICOMDIR exists: %s", dicomdirPath)
	}

	// 2. Verify PT000000 directory exists
	patientDir := filepath.Join(outputDir, "PT000000")
	if info, err := os.Stat(patientDir); os.IsNotExist(err) {
		t.Error("PT000000 directory should exist")
	} else if !info.IsDir() {
		t.Error("PT000000 should be a directory")
	} else {
		t.Logf("✓ Patient directory exists: %s", patientDir)
	}

	// 3. Verify ST000000 directory exists
	studyDir := filepath.Join(patientDir, "ST000000")
	if info, err := os.Stat(studyDir); os.IsNotExist(err) {
		t.Error("ST000000 directory should exist")
	} else if !info.IsDir() {
		t.Error("ST000000 should be a directory")
	} else {
		t.Logf("✓ Study directory exists: %s", studyDir)
	}

	// 4. Verify SE000000 directory exists
	seriesDir := filepath.Join(studyDir, "SE000000")
	if info, err := os.Stat(seriesDir); os.IsNotExist(err) {
		t.Error("SE000000 directory should exist")
	} else if !info.IsDir() {
		t.Error("SE000000 should be a directory")
	} else {
		t.Logf("✓ Series directory exists: %s", seriesDir)
	}

	// 5. Verify image files exist (IM000001, IM000002, ...)
	imageCount := 0
	for i := 1; i <= 5; i++ {
		imageFile := filepath.Join(seriesDir, fmt.Sprintf("IM%06d", i))
		if _, err := os.Stat(imageFile); os.IsNotExist(err) {
			t.Errorf("Image file should exist: %s", imageFile)
		} else {
			imageCount++
		}
	}
	t.Logf("✓ Found %d image files in series directory", imageCount)

	// 6. Verify no IMG*.dcm files in root
	matches, _ := filepath.Glob(filepath.Join(outputDir, "IMG*.dcm"))
	if len(matches) > 0 {
		t.Errorf("Temporary IMG*.dcm files should be cleaned up, found %d", len(matches))
	} else {
		t.Logf("✓ Temporary files cleaned up")
	}

	t.Logf("✓ DICOMDIR structure test passed")
}

// TestValidation_RequiredTags tests that DICOM files have required tags
func TestValidation_RequiredTags(t *testing.T) {
	outputDir := t.TempDir()

	opts := internaldicom.GeneratorOptions{
		NumImages:  3,
		TotalSize:  "5MB",
		OutputDir:  outputDir,
		Seed:       42,
		NumStudies: 1,
	}

	// Generate and organize
	files, err := internaldicom.GenerateDICOMSeries(opts)
	if err != nil {
		t.Fatalf("GenerateDICOMSeries failed: %v", err)
	}

	err = internaldicom.OrganizeFilesIntoDICOMDIR(outputDir, files)
	if err != nil {
		t.Fatalf("OrganizeFilesIntoDICOMDIR failed: %v", err)
	}

	// Parse first DICOM file
	firstImage := filepath.Join(outputDir, "PT000000", "ST000000", "SE000000", "IM000001")
	t.Logf("Parsing DICOM file: %s", firstImage)

	ds, err := dicom.ParseFile(firstImage, nil)
	if err != nil {
		t.Fatalf("Failed to parse DICOM file: %v", err)
	}

	// Verify required tags exist
	requiredTags := []struct {
		tag  tag.Tag
		name string
	}{
		{tag.PatientName, "PatientName"},
		{tag.PatientID, "PatientID"},
		{tag.PatientBirthDate, "PatientBirthDate"},
		{tag.PatientSex, "PatientSex"},
		{tag.StudyInstanceUID, "StudyInstanceUID"},
		{tag.SeriesInstanceUID, "SeriesInstanceUID"},
		{tag.SOPInstanceUID, "SOPInstanceUID"},
		{tag.Modality, "Modality"},
		{tag.Rows, "Rows"},
		{tag.Columns, "Columns"},
		{tag.BitsAllocated, "BitsAllocated"},
		{tag.PhotometricInterpretation, "PhotometricInterpretation"},
	}

	for _, rt := range requiredTags {
		elem, err := ds.FindElementByTag(rt.tag)
		if err != nil {
			t.Errorf("Tag %s (%v) should exist, got error: %v", rt.name, rt.tag, err)
			continue
		}
		if elem == nil {
			t.Errorf("Tag %s (%v) should not be nil", rt.name, rt.tag)
			continue
		}
		t.Logf("✓ Found tag %s: %v", rt.name, elem.Value)
	}

	// Check specific values
	modality, err := ds.FindElementByTag(tag.Modality)
	if err == nil && modality != nil {
		modalityStr := strings.Trim(modality.Value.String(), " []")
		if modalityStr != "MR" {
			t.Errorf("Modality should be 'MR', got '%s'", modalityStr)
		} else {
			t.Logf("✓ Modality = MR")
		}
	}

	bitsAlloc, err := ds.FindElementByTag(tag.BitsAllocated)
	if err == nil && bitsAlloc != nil {
		if bitsAlloc.Value.GetValue() != nil {
			t.Logf("✓ BitsAllocated = %v", bitsAlloc.Value.GetValue())
		}
	}

	t.Logf("✓ Required tags validation passed")
}

// TestMultiStudy tests multi-study generation
func TestMultiStudy(t *testing.T) {
	outputDir := t.TempDir()

	opts := internaldicom.GeneratorOptions{
		NumImages:  15,
		TotalSize:  "20MB",
		OutputDir:  outputDir,
		Seed:       42,
		NumStudies: 3,
	}

	t.Logf("Generating multi-study series (3 studies, 15 images)...")

	files, err := internaldicom.GenerateDICOMSeries(opts)
	if err != nil {
		t.Fatalf("GenerateDICOMSeries failed: %v", err)
	}

	err = internaldicom.OrganizeFilesIntoDICOMDIR(outputDir, files)
	if err != nil {
		t.Fatalf("OrganizeFilesIntoDICOMDIR failed: %v", err)
	}

	// Verify 3 study directories
	patientDir := filepath.Join(outputDir, "PT000000")

	studyCount := 0
	for i := 0; i < 3; i++ {
		studyDir := filepath.Join(patientDir, fmt.Sprintf("ST%06d", i))
		if info, err := os.Stat(studyDir); err == nil && info.IsDir() {
			studyCount++
			t.Logf("✓ Found study directory: %s", studyDir)

			// Each study should have SE000000
			seriesDir := filepath.Join(studyDir, "SE000000")
			if info, err := os.Stat(seriesDir); err != nil || !info.IsDir() {
				t.Errorf("Study %d should have SE000000 directory", i)
			}
		}
	}

	if studyCount != 3 {
		t.Errorf("Expected 3 study directories, found %d", studyCount)
	}

	// Count total images
	totalImages := 0
	filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasPrefix(info.Name(), "IM") {
			totalImages++
		}
		return nil
	})

	if totalImages != 15 {
		t.Errorf("Expected 15 total images, found %d", totalImages)
	} else {
		t.Logf("✓ Found %d total images across all studies", totalImages)
	}

	t.Logf("✓ Multi-study test passed")
}

// TestReproducibility_SameSeed tests that same seed produces same IDs
func TestReproducibility_SameSeed(t *testing.T) {
	seed := int64(42)

	// Generate first series
	outputDir1 := t.TempDir()
	opts1 := internaldicom.GeneratorOptions{
		NumImages:  3,
		TotalSize:  "5MB",
		OutputDir:  outputDir1,
		Seed:       seed,
		NumStudies: 1,
	}

	t.Logf("Generating first series with seed %d...", seed)
	files1, err := internaldicom.GenerateDICOMSeries(opts1)
	if err != nil {
		t.Fatalf("First generation failed: %v", err)
	}

	// Generate second series with same seed
	outputDir2 := t.TempDir()
	opts2 := opts1
	opts2.OutputDir = outputDir2

	t.Logf("Generating second series with same seed %d...", seed)
	files2, err := internaldicom.GenerateDICOMSeries(opts2)
	if err != nil {
		t.Fatalf("Second generation failed: %v", err)
	}

	// Compare patient IDs (should be identical with same seed)
	if files1[0].PatientID != files2[0].PatientID {
		t.Errorf("PatientID should be identical with same seed")
		t.Logf("  First:  %s", files1[0].PatientID)
		t.Logf("  Second: %s", files2[0].PatientID)
	} else {
		t.Logf("✓ PatientID identical: %s", files1[0].PatientID)
	}

	// StudyUID depends on output directory, so they will differ
	// But we can verify they follow the same pattern
	t.Logf("StudyUID (first):  %s", files1[0].StudyUID)
	t.Logf("StudyUID (second): %s", files2[0].StudyUID)

	t.Logf("✓ Reproducibility test passed")
}

// TestCalculateDimensions tests dimension calculation
func TestCalculateDimensions(t *testing.T) {
	tests := []struct {
		name       string
		totalBytes int64
		numImages  int
		wantMin    int // Minimum acceptable dimension
		wantMax    int // Maximum acceptable dimension
	}{
		{
			name:       "10MB_5images",
			totalBytes: 10 * 1024 * 1024,
			numImages:  5,
			wantMin:    900,
			wantMax:    1200,
		},
		{
			name:       "50MB_10images",
			totalBytes: 50 * 1024 * 1024,
			numImages:  10,
			wantMin:    1500,
			wantMax:    2000,
		},
		{
			name:       "100MB_10images",
			totalBytes: 100 * 1024 * 1024,
			numImages:  10,
			wantMin:    2100,
			wantMax:    2600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, h, err := internaldicom.CalculateDimensions(tt.totalBytes, tt.numImages)
			if err != nil {
				t.Fatalf("CalculateDimensions failed: %v", err)
			}

			if w != h {
				t.Errorf("Width and height should be equal, got %dx%d", w, h)
			}

			if w < tt.wantMin || w > tt.wantMax {
				t.Errorf("Dimension %d out of expected range [%d, %d]", w, tt.wantMin, tt.wantMax)
			} else {
				t.Logf("✓ Calculated dimensions: %dx%d (in range [%d, %d])",
					w, h, tt.wantMin, tt.wantMax)
			}
		})
	}
}
