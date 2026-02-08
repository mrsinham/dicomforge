package corruption

import (
	"math/rand/v2"
	"testing"
)

func TestGeneratePhilipsPrivateElements(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))
	elements := generatePhilipsPrivateElements(rng)

	if len(elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(elements))
	}

	// Verify Philips Imaging DD 001 creator
	if elements[0].Tag.Group != 0x2001 || elements[0].Tag.Element != 0x0010 {
		t.Errorf("first element should be (2001,0010), got %v", elements[0].Tag)
	}

	// Verify Philips MR Imaging DD 001 creator
	if elements[1].Tag.Group != 0x2005 || elements[1].Tag.Element != 0x0010 {
		t.Errorf("second element should be (2005,0010), got %v", elements[1].Tag)
	}

	// Verify private sequence
	if elements[2].Tag.Group != 0x2005 || elements[2].Tag.Element != 0x100E {
		t.Errorf("third element should be (2005,100E), got %v", elements[2].Tag)
	}

	// Verify it's a sequence (SQ VR)
	if elements[2].RawValueRepresentation != "SQ" {
		t.Errorf("third element should have SQ VR, got %s", elements[2].RawValueRepresentation)
	}
}
