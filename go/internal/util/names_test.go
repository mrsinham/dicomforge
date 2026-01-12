package util

import (
	"math/rand/v2"
	"strings"
	"testing"
)

func TestGeneratePatientName_Format(t *testing.T) {
	name := GeneratePatientName("M", nil)

	if !strings.Contains(name, "^") {
		t.Errorf("Name should contain '^' separator, got: %s", name)
	}

	parts := strings.Split(name, "^")
	if len(parts) != 2 {
		t.Errorf("Name should have exactly 2 parts (LASTNAME^FIRSTNAME), got: %s", name)
	}
}

func TestGeneratePatientName_Deterministic(t *testing.T) {
	// Test that same seed produces same name
	source1 := rand.NewPCG(42, 42)
	rng1 := rand.New(source1)
	name1 := GeneratePatientName("M", rng1)

	source2 := rand.NewPCG(42, 42)
	rng2 := rand.New(source2)
	name2 := GeneratePatientName("M", rng2)

	if name1 != name2 {
		t.Errorf("Same seed should produce same name: %s != %s", name1, name2)
	}

	// Also test female names
	source3 := rand.NewPCG(99, 99)
	rng3 := rand.New(source3)
	femaleName1 := GeneratePatientName("F", rng3)

	source4 := rand.NewPCG(99, 99)
	rng4 := rand.New(source4)
	femaleName2 := GeneratePatientName("F", rng4)

	if femaleName1 != femaleName2 {
		t.Errorf("Same seed should produce same name: %s != %s", femaleName1, femaleName2)
	}
}

func TestGeneratePatientName_Sex(t *testing.T) {
	// Test male names
	for i := 0; i < 10; i++ {
		name := GeneratePatientName("M", nil)
		parts := strings.Split(name, "^")
		firstName := parts[1]

		found := false
		for _, mn := range MaleFirstNames {
			if mn == firstName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Male name %s not in male first names list", firstName)
		}
	}

	// Test female names
	for i := 0; i < 10; i++ {
		name := GeneratePatientName("F", nil)
		parts := strings.Split(name, "^")
		firstName := parts[1]

		found := false
		for _, fn := range FemaleFirstNames {
			if fn == firstName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Female name %s not in female first names list", firstName)
		}
	}
}
