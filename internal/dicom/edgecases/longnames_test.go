package edgecases

import (
	"math/rand/v2"
	"strings"
	"testing"
)

func TestGenerateLongPatientName(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))
	name := GenerateLongPatientName("M", rng)
	if len(name) < 50 {
		t.Errorf("Long name should be >= 50 chars, got %d: %s", len(name), name)
	}
	if len(name) > 64 {
		t.Errorf("Long name should be <= 64 chars (DICOM LO max), got %d", len(name))
	}
	if !strings.Contains(name, "^") {
		t.Errorf("Name should have DICOM format with ^: %s", name)
	}
}

func TestGenerateLongPatientID(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))
	id := GenerateLongPatientID(rng)
	if len(id) != 64 {
		t.Errorf("Long ID should be exactly 64 chars, got %d", len(id))
	}
}

func TestGenerateLongStudyDescription(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))
	desc := GenerateLongStudyDescription(rng)
	if len(desc) < 50 || len(desc) > 64 {
		t.Errorf("Long description should be 50-64 chars, got %d", len(desc))
	}
}
