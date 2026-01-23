package edgecases

import (
	"math/rand/v2"
	"strings"
	"testing"
	"unicode"
)

func hasSpecialChar(s string) bool {
	for _, r := range s {
		if r == '-' || r == '\'' || r > 127 || !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '^' && r != ' ' {
			return true
		}
	}
	return false
}

func TestGenerateSpecialCharName(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))
	for i := 0; i < 10; i++ {
		name := GenerateSpecialCharName("M", rng)
		if !strings.Contains(name, "^") {
			t.Errorf("Name should have DICOM format with ^: %s", name)
		}
		if !hasSpecialChar(name) {
			t.Errorf("Name should contain special character: %s", name)
		}
	}
}

func TestGenerateSpecialCharName_Deterministic(t *testing.T) {
	name1 := GenerateSpecialCharName("F", rand.New(rand.NewPCG(42, 42)))
	name2 := GenerateSpecialCharName("F", rand.New(rand.NewPCG(42, 42)))
	if name1 != name2 {
		t.Errorf("Same seed should produce same name: %s vs %s", name1, name2)
	}
}
