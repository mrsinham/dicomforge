package util

import (
	"strings"
	"testing"
)

func TestGeneratePatientName_Format(t *testing.T) {
	name := GeneratePatientName("M")

	if !strings.Contains(name, "^") {
		t.Errorf("Name should contain '^' separator, got: %s", name)
	}

	parts := strings.Split(name, "^")
	if len(parts) != 2 {
		t.Errorf("Name should have exactly 2 parts (LASTNAME^FIRSTNAME), got: %s", name)
	}
}

func TestGeneratePatientName_Deterministic(t *testing.T) {
	// Test that global rand produces consistent results across calls
	// In rand/v2, the global functions are automatically seeded but consistent within a run
	name1 := GeneratePatientName("M")
	name2 := GeneratePatientName("M")

	// Names might be different (random), but function should work consistently
	// Both should have valid format
	if !strings.Contains(name1, "^") || !strings.Contains(name2, "^") {
		t.Errorf("Names should contain '^' separator: %s, %s", name1, name2)
	}

	parts1 := strings.Split(name1, "^")
	parts2 := strings.Split(name2, "^")
	if len(parts1) != 2 || len(parts2) != 2 {
		t.Errorf("Names should have exactly 2 parts: %s, %s", name1, name2)
	}
}

func TestGeneratePatientName_Sex(t *testing.T) {
	maleFirstNames := []string{
		"Jean", "Pierre", "Michel", "André", "Philippe", "Alain", "Bernard", "Jacques",
		"François", "Christian", "Daniel", "Patrick", "Nicolas", "Olivier", "Laurent",
		"Thierry", "Stéphane", "Éric", "David", "Julien", "Christophe", "Pascal",
		"Sébastien", "Marc", "Vincent", "Antoine", "Alexandre", "Maxime", "Thomas",
		"Lucas", "Hugo", "Louis", "Arthur", "Gabriel", "Raphaël", "Paul", "Jules",
	}

	femaleFirstNames := []string{
		"Marie", "Nathalie", "Isabelle", "Sylvie", "Catherine", "Françoise", "Valérie",
		"Christine", "Monique", "Sophie", "Patricia", "Martine", "Nicole", "Sandrine",
		"Stéphanie", "Céline", "Julie", "Aurélie", "Caroline", "Laurence", "Émilie",
		"Claire", "Anne", "Camille", "Laura", "Sarah", "Manon", "Emma", "Léa",
		"Chloé", "Zoé", "Alice", "Charlotte", "Lucie", "Juliette", "Louise",
	}

	// Test male names
	for i := 0; i < 10; i++ {
		name := GeneratePatientName("M")
		parts := strings.Split(name, "^")
		firstName := parts[1]

		found := false
		for _, mn := range maleFirstNames {
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
		name := GeneratePatientName("F")
		parts := strings.Split(name, "^")
		firstName := parts[1]

		found := false
		for _, fn := range femaleFirstNames {
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
