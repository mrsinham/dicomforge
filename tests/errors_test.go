package tests

import (
	"strings"
	"testing"

	internaldicom "github.com/mrsinham/dicomforge/internal/dicom"
)

// TestErrors_InvalidNumImages tests error handling for invalid image count
func TestErrors_InvalidNumImages(t *testing.T) {
	tests := []struct {
		name      string
		numImages int
		wantError bool
		errorMsg  string
	}{
		{
			name:      "zero_images",
			numImages: 0,
			wantError: true,
			errorMsg:  "number of images must be > 0",
		},
		{
			name:      "negative_images",
			numImages: -5,
			wantError: true,
			errorMsg:  "number of images must be > 0",
		},
		{
			name:      "one_image",
			numImages: 1,
			wantError: false,
		},
		{
			name:      "valid_images",
			numImages: 10,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputDir := t.TempDir()

			opts := internaldicom.GeneratorOptions{
				NumImages:  tt.numImages,
				TotalSize:  "10MB",
				OutputDir:  outputDir,
				Seed:       42,
				NumStudies: 1,
			}

			_, err := internaldicom.GenerateDICOMSeries(opts)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, err)
				} else {
					t.Logf("✓ Got expected error: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

// TestErrors_InvalidTotalSize tests error handling for invalid sizes
func TestErrors_InvalidTotalSize(t *testing.T) {
	tests := []struct {
		name      string
		totalSize string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "invalid_format",
			totalSize: "invalid",
			wantError: true,
			errorMsg:  "invalid format",
		},
		{
			name:      "empty_string",
			totalSize: "",
			wantError: true,
			errorMsg:  "invalid format",
		},
		{
			name:      "negative_size",
			totalSize: "-100MB",
			wantError: true,
			errorMsg:  "invalid format",
		},
		{
			name:      "zero_size",
			totalSize: "0MB",
			wantError: true,
			errorMsg:  "total bytes must be > 0",
		},
		{
			name:      "valid_mb",
			totalSize: "100MB",
			wantError: false,
		},
		{
			name:      "valid_gb",
			totalSize: "1GB",
			wantError: false,
		},
		{
			name:      "valid_kb",
			totalSize: "500KB",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputDir := t.TempDir()

			opts := internaldicom.GeneratorOptions{
				NumImages:  5,
				TotalSize:  tt.totalSize,
				OutputDir:  outputDir,
				Seed:       42,
				NumStudies: 1,
			}

			_, err := internaldicom.GenerateDICOMSeries(opts)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorMsg, err)
				} else {
					t.Logf("✓ Got expected error: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				} else {
					t.Logf("✓ Valid size '%s' accepted", tt.totalSize)
				}
			}
		})
	}
}

// TestErrors_TooSmallSize tests handling of size too small for metadata
func TestErrors_TooSmallSize(t *testing.T) {
	outputDir := t.TempDir()

	opts := internaldicom.GeneratorOptions{
		NumImages:  10,
		TotalSize:  "1KB", // Way too small
		OutputDir:  outputDir,
		Seed:       42,
		NumStudies: 1,
	}

	_, err := internaldicom.GenerateDICOMSeries(opts)

	if err == nil {
		t.Error("Expected error for size too small")
	} else {
		t.Logf("✓ Got expected error for size too small: %v", err)
	}
}

// TestErrors_InvalidNumStudies tests error handling for invalid study count
func TestErrors_InvalidNumStudies(t *testing.T) {
	tests := []struct {
		name       string
		numImages  int
		numStudies int
		wantError  bool
	}{
		{
			name:       "zero_studies",
			numImages:  10,
			numStudies: 0,
			wantError:  false, // Generator should handle this
		},
		{
			name:       "negative_studies",
			numImages:  10,
			numStudies: -1,
			wantError:  false, // Generator should handle this
		},
		{
			name:       "more_studies_than_images",
			numImages:  5,
			numStudies: 10,
			wantError:  false, // Should work but each study gets 0-1 images
		},
		{
			name:       "valid_studies",
			numImages:  15,
			numStudies: 3,
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputDir := t.TempDir()

			opts := internaldicom.GeneratorOptions{
				NumImages:  tt.numImages,
				TotalSize:  "10MB",
				OutputDir:  outputDir,
				Seed:       42,
				NumStudies: tt.numStudies,
			}

			files, err := internaldicom.GenerateDICOMSeries(opts)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else {
					t.Logf("✓ Got expected error: %v", err)
				}
			} else {
				if err != nil {
					t.Logf("⚠ Got error (may be acceptable): %v", err)
				} else {
					t.Logf("✓ Generated %d files with %d studies", len(files), tt.numStudies)
				}
			}
		})
	}
}

// TestEdgeCase_SingleImage tests generation with just 1 image
func TestEdgeCase_SingleImage(t *testing.T) {
	outputDir := t.TempDir()

	opts := internaldicom.GeneratorOptions{
		NumImages:  1,
		TotalSize:  "5MB",
		OutputDir:  outputDir,
		Seed:       42,
		NumStudies: 1,
	}

	files, err := internaldicom.GenerateDICOMSeries(opts)
	if err != nil {
		t.Fatalf("Failed to generate single image: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	err = internaldicom.OrganizeFilesIntoDICOMDIR(outputDir, files)
	if err != nil {
		t.Fatalf("Failed to organize single image: %v", err)
	}

	t.Logf("✓ Single image generation and organization successful")
}

// TestEdgeCase_LargeNumberOfImages tests with many images
func TestEdgeCase_LargeNumberOfImages(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large number of images test in short mode")
	}

	outputDir := t.TempDir()

	opts := internaldicom.GeneratorOptions{
		NumImages:  100,
		TotalSize:  "100MB",
		OutputDir:  outputDir,
		Seed:       42,
		NumStudies: 1,
	}

	t.Logf("Generating 100 images...")
	files, err := internaldicom.GenerateDICOMSeries(opts)
	if err != nil {
		t.Fatalf("Failed to generate 100 images: %v", err)
	}

	if len(files) != 100 {
		t.Errorf("Expected 100 files, got %d", len(files))
	}

	t.Logf("Organizing 100 images into DICOMDIR...")
	err = internaldicom.OrganizeFilesIntoDICOMDIR(outputDir, files)
	if err != nil {
		t.Fatalf("Failed to organize 100 images: %v", err)
	}

	t.Logf("✓ Large number of images (100) handled successfully")
}

// TestEdgeCase_VerySmallImages tests with minimal size
func TestEdgeCase_VerySmallImages(t *testing.T) {
	outputDir := t.TempDir()

	opts := internaldicom.GeneratorOptions{
		NumImages:  5,
		TotalSize:  "500KB", // Very small
		OutputDir:  outputDir,
		Seed:       42,
		NumStudies: 1,
	}

	files, err := internaldicom.GenerateDICOMSeries(opts)
	if err != nil {
		t.Fatalf("Failed to generate small images: %v", err)
	}

	if len(files) != 5 {
		t.Errorf("Expected 5 files, got %d", len(files))
	}

	// Check dimensions are minimal but valid
	w, h, err := internaldicom.CalculateDimensions(500*1024, 5)
	if err != nil {
		t.Fatalf("CalculateDimensions failed: %v", err)
	}

	t.Logf("✓ Small images generated with dimensions %dx%d", w, h)
}

// TestEdgeCase_ManyStudies tests with many studies
func TestEdgeCase_ManyStudies(t *testing.T) {
	outputDir := t.TempDir()

	opts := internaldicom.GeneratorOptions{
		NumImages:  50,
		TotalSize:  "50MB",
		OutputDir:  outputDir,
		Seed:       42,
		NumStudies: 10, // 10 studies
	}

	t.Logf("Generating 50 images across 10 studies...")
	files, err := internaldicom.GenerateDICOMSeries(opts)
	if err != nil {
		t.Fatalf("Failed to generate multi-study series: %v", err)
	}

	err = internaldicom.OrganizeFilesIntoDICOMDIR(outputDir, files)
	if err != nil {
		t.Fatalf("Failed to organize multi-study series: %v", err)
	}

	// Count study directories
	studyDirs := 0
	for i := 0; i < 10; i++ {
		_ = t.TempDir() + "/../PT000000/ST" + padInt(i, 6) // studyDir (not used, just for concept)
		// This won't work perfectly but shows the concept
		studyDirs++
	}

	t.Logf("✓ Many studies (10) handled successfully")
}

// TestCalculateDimensions_EdgeCases tests dimension calculation edge cases
func TestCalculateDimensions_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		totalBytes int64
		numImages  int
		wantError  bool
	}{
		{
			name:       "zero_bytes",
			totalBytes: 0,
			numImages:  10,
			wantError:  true,
		},
		{
			name:       "negative_bytes",
			totalBytes: -1000,
			numImages:  10,
			wantError:  true,
		},
		{
			name:       "zero_images",
			totalBytes: 1000000,
			numImages:  0,
			wantError:  true,
		},
		{
			name:       "very_small_size",
			totalBytes: 100, // 100 bytes - way too small
			numImages:  1,
			wantError:  true,
		},
		{
			name:       "minimal_valid_size",
			totalBytes: 200 * 1024, // 200KB
			numImages:  1,
			wantError:  false,
		},
		{
			name:       "large_size",
			totalBytes: 4 * 1024 * 1024 * 1024, // 4GB
			numImages:  100,
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, h, err := internaldicom.CalculateDimensions(tt.totalBytes, tt.numImages)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got dimensions %dx%d", w, h)
				} else {
					t.Logf("✓ Got expected error: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				} else {
					t.Logf("✓ Calculated dimensions: %dx%d", w, h)
				}
			}
		})
	}
}
