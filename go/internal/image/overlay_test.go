package image

import (
	"testing"
)

func TestAddTextOverlay_Range(t *testing.T) {
	width, height := 256, 256
	pixels := GenerateSingleImage(width, height, 42)

	err := AddTextOverlay(pixels, width, height, 5, 10)
	if err != nil {
		t.Fatalf("AddTextOverlay failed: %v", err)
	}

	// Verify pixels still in valid range
	for i, pixel := range pixels {
		if pixel > 4095 {
			t.Errorf("Pixel %d value %d exceeds 12-bit max after overlay", i, pixel)
		}
	}
}

func TestAddTextOverlay_ModifiesImage(t *testing.T) {
	width, height := 256, 256
	pixels := GenerateSingleImage(width, height, 42)

	// Make a copy before overlay
	original := make([]uint16, len(pixels))
	copy(original, pixels)

	err := AddTextOverlay(pixels, width, height, 5, 10)
	if err != nil {
		t.Fatalf("AddTextOverlay failed: %v", err)
	}

	// Check that at least some pixels changed (text was drawn)
	different := false
	for i := range pixels {
		if pixels[i] != original[i] {
			different = true
			break
		}
	}

	if !different {
		t.Errorf("Expected overlay to modify pixels")
	}
}

func TestAddTextOverlay_Validation(t *testing.T) {
	tests := []struct {
		name          string
		width         int
		height        int
		imageNum      int
		totalImages   int
		pixelsLen     int
		expectError   bool
		errorContains string
	}{
		{
			name:          "invalid dimensions - negative width",
			width:         -10,
			height:        100,
			imageNum:      1,
			totalImages:   10,
			pixelsLen:     1000,
			expectError:   true,
			errorContains: "invalid dimensions",
		},
		{
			name:          "invalid dimensions - zero height",
			width:         100,
			height:        0,
			imageNum:      1,
			totalImages:   10,
			pixelsLen:     0,
			expectError:   true,
			errorContains: "invalid dimensions",
		},
		{
			name:          "pixel count mismatch",
			width:         100,
			height:        100,
			imageNum:      1,
			totalImages:   10,
			pixelsLen:     500,
			expectError:   true,
			errorContains: "does not match dimensions",
		},
		{
			name:          "invalid numbering - imageNum zero",
			width:         100,
			height:        100,
			imageNum:      0,
			totalImages:   10,
			pixelsLen:     10000,
			expectError:   true,
			errorContains: "invalid image numbering",
		},
		{
			name:          "invalid numbering - totalImages zero",
			width:         100,
			height:        100,
			imageNum:      1,
			totalImages:   0,
			pixelsLen:     10000,
			expectError:   true,
			errorContains: "invalid image numbering",
		},
		{
			name:          "invalid numbering - imageNum exceeds totalImages",
			width:         100,
			height:        100,
			imageNum:      150,
			totalImages:   100,
			pixelsLen:     10000,
			expectError:   true,
			errorContains: "invalid image numbering",
		},
		{
			name:        "valid input",
			width:       100,
			height:      100,
			imageNum:    5,
			totalImages: 10,
			pixelsLen:   10000,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pixels := make([]uint16, tt.pixelsLen)
			err := AddTextOverlay(pixels, tt.width, tt.height, tt.imageNum, tt.totalImages)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorContains)
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
