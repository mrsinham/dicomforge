// internal/util/priority_test.go
package util

import (
	"math/rand/v2"
	"testing"
)

func TestParsePriority_Valid(t *testing.T) {
	tests := []struct {
		input    string
		expected Priority
	}{
		{"HIGH", PriorityHigh},
		{"high", PriorityHigh},
		{"ROUTINE", PriorityRoutine},
		{"routine", PriorityRoutine},
		{"LOW", PriorityLow},
		{"low", PriorityLow},
	}

	for _, tc := range tests {
		result, err := ParsePriority(tc.input)
		if err != nil {
			t.Errorf("ParsePriority(%q) returned error: %v", tc.input, err)
		}
		if result != tc.expected {
			t.Errorf("ParsePriority(%q) = %v, want %v", tc.input, result, tc.expected)
		}
	}
}

func TestParsePriority_Invalid(t *testing.T) {
	_, err := ParsePriority("INVALID")
	if err == nil {
		t.Error("ParsePriority(INVALID) should return error")
	}
}

func TestPriority_String(t *testing.T) {
	if PriorityHigh.String() != "HIGH" {
		t.Errorf("PriorityHigh.String() = %s, want HIGH", PriorityHigh.String())
	}
	if PriorityRoutine.String() != "ROUTINE" {
		t.Errorf("PriorityRoutine.String() = %s, want ROUTINE", PriorityRoutine.String())
	}
	if PriorityLow.String() != "LOW" {
		t.Errorf("PriorityLow.String() = %s, want LOW", PriorityLow.String())
	}
}

func TestGeneratePriority_Distribution(t *testing.T) {
	// Generate many priorities and check distribution
	counts := map[Priority]int{}
	rng := rand.New(rand.NewPCG(42, 42))

	for i := 0; i < 1000; i++ {
		p := GeneratePriority(rng)
		counts[p]++
	}

	// ROUTINE should be most common (~70%), HIGH ~20%, LOW ~10%
	if counts[PriorityRoutine] < 500 {
		t.Errorf("ROUTINE should be most common, got %d/1000", counts[PriorityRoutine])
	}
}
