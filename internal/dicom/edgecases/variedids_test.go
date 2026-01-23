package edgecases

import (
	"math/rand/v2"
	"strings"
	"testing"
)

func TestGenerateVariedPatientID_WithDashes(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))
	id := GenerateVariedPatientID(IDWithDashes, rng)
	if !strings.Contains(id, "-") {
		t.Errorf("ID with dashes should contain '-': %s", id)
	}
}

func TestGenerateVariedPatientID_WithLetters(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))
	id := GenerateVariedPatientID(IDWithLetters, rng)
	hasLetter := false
	hasDigit := false
	for _, c := range id {
		if c >= 'A' && c <= 'Z' {
			hasLetter = true
		}
		if c >= '0' && c <= '9' {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		t.Errorf("ID with letters should have both letters and digits: %s", id)
	}
}

func TestGenerateVariedPatientID_WithSpaces(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))
	id := GenerateVariedPatientID(IDWithSpaces, rng)
	if !strings.Contains(id, " ") {
		t.Errorf("ID with spaces should contain space: %s", id)
	}
}

func TestGenerateVariedPatientID_Long(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))
	id := GenerateVariedPatientID(IDLong, rng)
	if len(id) != 64 {
		t.Errorf("Long ID should be 64 chars, got %d", len(id))
	}
}
