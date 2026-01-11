package util

import "testing"

func TestParseSize_ValidSizes(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"100KB", 102400},
		{"1MB", 1048576},
		{"1.5GB", 1610612736},
		{"500MB", 524288000},
		{"0.5KB", 512},
		{"0KB", 0},
		{"0MB", 0},
		{"0GB", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseSize(tt.input)
			if err != nil {
				t.Fatalf("ParseSize(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.expected {
				t.Errorf("ParseSize(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseSize_InvalidFormats(t *testing.T) {
	tests := []string{
		"100",
		"1.5TB",
		"abc",
		"100 MB",
		"-100MB",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParseSize(input)
			if err == nil {
				t.Errorf("ParseSize(%q) expected error, got nil", input)
			}
		})
	}
}
