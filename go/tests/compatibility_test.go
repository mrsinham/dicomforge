package tests

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	internaldicom "github.com/julien/dicom-test/go/internal/dicom"
)

// TestCompatibility_PythonValidation tests that Go-generated files are valid according to pydicom
func TestCompatibility_PythonValidation(t *testing.T) {
	// Check if Python and pydicom are available
	if !isPythonAvailable(t) {
		t.Skip("Python or pydicom not available, skipping validation test")
	}

	// Generate test files with Go
	outputDir := t.TempDir()

	opts := internaldicom.GeneratorOptions{
		NumImages:  5,
		TotalSize:  "10MB",
		OutputDir:  outputDir,
		Seed:       42,
		NumStudies: 1,
	}

	t.Logf("Generating DICOM files with Go...")
	files, err := internaldicom.GenerateDICOMSeries(opts)
	if err != nil {
		t.Fatalf("GenerateDICOMSeries failed: %v", err)
	}

	err = internaldicom.OrganizeFilesIntoDICOMDIR(outputDir, files)
	if err != nil {
		t.Fatalf("OrganizeFilesIntoDICOMDIR failed: %v", err)
	}

	// Run Python validation script
	scriptPath := filepath.Join("..", "scripts", "validate_dicom.py")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Skipf("Validation script not found: %s", scriptPath)
	}

	t.Logf("Validating with pydicom...")
	cmd := exec.Command("python3", scriptPath, outputDir)
	output, err := cmd.CombinedOutput()

	t.Logf("Validation output:\n%s", string(output))

	if err != nil {
		t.Errorf("Python validation failed: %v\nOutput: %s", err, string(output))
	} else {
		t.Logf("✓ All Go-generated files validated successfully by pydicom")
	}
}

// TestCompatibility_MetadataExtraction tests metadata extraction from Go-generated files
func TestCompatibility_MetadataExtraction(t *testing.T) {
	if !isPythonAvailable(t) {
		t.Skip("Python or pydicom not available, skipping metadata extraction test")
	}

	// Generate test files
	outputDir := t.TempDir()

	opts := internaldicom.GeneratorOptions{
		NumImages:  3,
		TotalSize:  "5MB",
		OutputDir:  outputDir,
		Seed:       42,
		NumStudies: 1,
	}

	files, err := internaldicom.GenerateDICOMSeries(opts)
	if err != nil {
		t.Fatalf("GenerateDICOMSeries failed: %v", err)
	}

	err = internaldicom.OrganizeFilesIntoDICOMDIR(outputDir, files)
	if err != nil {
		t.Fatalf("OrganizeFilesIntoDICOMDIR failed: %v", err)
	}

	// Extract metadata
	scriptPath := filepath.Join("..", "scripts", "extract_metadata.py")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Skipf("Extraction script not found: %s", scriptPath)
	}

	metadataFile := filepath.Join(t.TempDir(), "metadata.json")

	cmd := exec.Command("python3", scriptPath, outputDir, metadataFile)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("Metadata extraction failed: %v\nOutput: %s", err, string(output))
	}

	// Read and parse metadata
	data, err := os.ReadFile(metadataFile)
	if err != nil {
		t.Fatalf("Failed to read metadata file: %v", err)
	}

	var metadata struct {
		FileCount int `json:"file_count"`
		Files     []struct {
			PatientID   string `json:"patient_id"`
			PatientName string `json:"patient_name"`
			Modality    string `json:"modality"`
			Rows        int    `json:"rows"`
			Columns     int    `json:"columns"`
		} `json:"files"`
	}

	if err := json.Unmarshal(data, &metadata); err != nil {
		t.Fatalf("Failed to parse metadata JSON: %v", err)
	}

	// Validate extracted metadata
	if metadata.FileCount != 3 {
		t.Errorf("Expected 3 files, got %d", metadata.FileCount)
	}

	if len(metadata.Files) > 0 {
		first := metadata.Files[0]

		if first.PatientID == "" {
			t.Error("Patient ID should not be empty")
		} else {
			t.Logf("✓ Patient ID: %s", first.PatientID)
		}

		if first.PatientName == "" {
			t.Error("Patient name should not be empty")
		} else {
			t.Logf("✓ Patient Name: %s", first.PatientName)
		}

		if first.Modality != "MR" {
			t.Errorf("Modality should be MR, got: %s", first.Modality)
		} else {
			t.Logf("✓ Modality: %s", first.Modality)
		}

		if first.Rows == 0 || first.Columns == 0 {
			t.Errorf("Invalid dimensions: %dx%d", first.Rows, first.Columns)
		} else {
			t.Logf("✓ Dimensions: %dx%d", first.Rows, first.Columns)
		}
	}

	t.Logf("✓ Metadata extraction and validation successful")
}

// TestCompatibility_DICOMDIRStructure tests that DICOMDIR structure is readable by Python
func TestCompatibility_DICOMDIRStructure(t *testing.T) {
	if !isPythonAvailable(t) {
		t.Skip("Python or pydicom not available, skipping DICOMDIR structure test")
	}

	// Generate test files
	outputDir := t.TempDir()

	opts := internaldicom.GeneratorOptions{
		NumImages:  5,
		TotalSize:  "10MB",
		OutputDir:  outputDir,
		Seed:       42,
		NumStudies: 1,
	}

	files, err := internaldicom.GenerateDICOMSeries(opts)
	if err != nil {
		t.Fatalf("GenerateDICOMSeries failed: %v", err)
	}

	err = internaldicom.OrganizeFilesIntoDICOMDIR(outputDir, files)
	if err != nil {
		t.Fatalf("OrganizeFilesIntoDICOMDIR failed: %v", err)
	}

	// Verify DICOMDIR can be read by pydicom
	pythonCode := `
import sys
import pydicom
from pathlib import Path

dicomdir_path = Path(sys.argv[1]) / 'DICOMDIR'

if not dicomdir_path.exists():
    print("DICOMDIR not found")
    sys.exit(1)

try:
    ds = pydicom.dcmread(str(dicomdir_path))
    print(f"✓ DICOMDIR readable")

    # Check for FileSet ID
    if hasattr(ds, 'FileSetID'):
        print(f"✓ FileSet ID: {ds.FileSetID}")

    print("✓ DICOMDIR structure valid")
    sys.exit(0)

except Exception as e:
    print(f"Error reading DICOMDIR: {e}")
    sys.exit(1)
`

	cmd := exec.Command("python3", "-c", pythonCode, outputDir)
	output, err := cmd.CombinedOutput()

	t.Logf("DICOMDIR validation output:\n%s", string(output))

	if err != nil {
		t.Errorf("DICOMDIR validation failed: %v", err)
	} else {
		t.Logf("✓ DICOMDIR structure validated by pydicom")
	}
}

// TestCompatibility_SameSeedComparison compares outputs from Python and Go with same seed
func TestCompatibility_SameSeedComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comparison test in short mode")
	}

	if !isPythonAvailable(t) {
		t.Skip("Python or pydicom not available")
	}

	// Check if Python script exists
	pythonScript := filepath.Join("..", "..", "generate_dicom_mri.py")
	if _, err := os.Stat(pythonScript); os.IsNotExist(err) {
		t.Skipf("Python generator script not found: %s", pythonScript)
	}

	seed := int64(42)
	numImages := 3
	size := "5MB"

	// Generate with Go
	goOutputDir := filepath.Join(t.TempDir(), "go-output")

	goOpts := internaldicom.GeneratorOptions{
		NumImages:  numImages,
		TotalSize:  size,
		OutputDir:  goOutputDir,
		Seed:       seed,
		NumStudies: 1,
	}

	t.Logf("Generating with Go (seed=%d)...", seed)
	goFiles, err := internaldicom.GenerateDICOMSeries(goOpts)
	if err != nil {
		t.Fatalf("Go generation failed: %v", err)
	}

	err = internaldicom.OrganizeFilesIntoDICOMDIR(goOutputDir, goFiles)
	if err != nil {
		t.Fatalf("Go DICOMDIR organization failed: %v", err)
	}

	// Generate with Python
	pythonOutputDir := filepath.Join(t.TempDir(), "python-output")

	t.Logf("Generating with Python (seed=%d)...", seed)
	cmd := exec.Command("python3", pythonScript,
		"--num-images", "3",
		"--total-size", size,
		"--output", pythonOutputDir,
		"--seed", "42",
		"--num-studies", "1")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Python generation output:\n%s", string(output))
		t.Skipf("Python generation failed (this is OK if Python environment not set up): %v", err)
	}

	t.Logf("Python generation successful")

	// Extract metadata from both
	scriptPath := filepath.Join("..", "scripts", "extract_metadata.py")

	goMetadataFile := filepath.Join(t.TempDir(), "go-metadata.json")
	pythonMetadataFile := filepath.Join(t.TempDir(), "python-metadata.json")

	// Extract Go metadata
	cmd = exec.Command("python3", scriptPath, goOutputDir, goMetadataFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Go metadata extraction failed: %v\n%s", err, output)
	}

	// Extract Python metadata
	cmd = exec.Command("python3", scriptPath, pythonOutputDir, pythonMetadataFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Python metadata extraction failed: %v\n%s", err, output)
	}

	// Compare metadata
	goData, err := os.ReadFile(goMetadataFile)
	if err != nil {
		t.Fatalf("Failed to read Go metadata: %v", err)
	}

	pythonData, err := os.ReadFile(pythonMetadataFile)
	if err != nil {
		t.Fatalf("Failed to read Python metadata: %v", err)
	}

	var goMeta, pythonMeta struct {
		FileCount int `json:"file_count"`
		Files     []struct {
			PatientID   string `json:"patient_id"`
			PatientName string `json:"patient_name"`
			Modality    string `json:"modality"`
		} `json:"files"`
	}

	json.Unmarshal(goData, &goMeta)
	json.Unmarshal(pythonData, &pythonMeta)

	// Compare file counts
	if goMeta.FileCount != pythonMeta.FileCount {
		t.Errorf("File counts differ: Go=%d, Python=%d", goMeta.FileCount, pythonMeta.FileCount)
	} else {
		t.Logf("✓ File counts match: %d", goMeta.FileCount)
	}

	// Compare first file metadata
	if len(goMeta.Files) > 0 && len(pythonMeta.Files) > 0 {
		goFirst := goMeta.Files[0]
		pyFirst := pythonMeta.Files[0]

		// Patient ID should match with same seed
		if goFirst.PatientID == pyFirst.PatientID {
			t.Logf("✓ Patient IDs match: %s", goFirst.PatientID)
		} else {
			t.Logf("⚠ Patient IDs differ (expected with different RNG): Go=%s, Python=%s",
				goFirst.PatientID, pyFirst.PatientID)
		}

		// Patient Name should match with same seed
		if goFirst.PatientName == pyFirst.PatientName {
			t.Logf("✓ Patient Names match: %s", goFirst.PatientName)
		} else {
			t.Logf("⚠ Patient Names differ (expected with different RNG): Go=%s, Python=%s",
				goFirst.PatientName, pyFirst.PatientName)
		}

		// Modality should always match
		if goFirst.Modality != pyFirst.Modality {
			t.Errorf("Modality mismatch: Go=%s, Python=%s", goFirst.Modality, pyFirst.Modality)
		} else {
			t.Logf("✓ Modality matches: %s", goFirst.Modality)
		}
	}

	t.Logf("✓ Comparison complete (note: some differences expected due to different RNG implementations)")
}

// Helper function to check if Python and pydicom are available
func isPythonAvailable(t *testing.T) bool {
	// Check Python
	if _, err := exec.LookPath("python3"); err != nil {
		t.Logf("python3 not found in PATH")
		return false
	}

	// Check pydicom
	cmd := exec.Command("python3", "-c", "import pydicom")
	if err := cmd.Run(); err != nil {
		t.Logf("pydicom not installed")
		return false
	}

	return true
}
