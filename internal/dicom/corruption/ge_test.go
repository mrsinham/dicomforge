package corruption

import (
	"math/rand/v2"
	"testing"
)

func TestGenerateGEPrivateElements(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))
	elements := generateGEPrivateElements(rng)

	if len(elements) != 4 {
		t.Fatalf("expected 4 elements, got %d", len(elements))
	}

	// Verify GEMS_IDEN_01 creator
	if elements[0].Tag.Group != 0x0009 || elements[0].Tag.Element != 0x0010 {
		t.Errorf("first element should be (0009,0010), got %v", elements[0].Tag)
	}

	// Verify GEMS_PARM_01 creator
	if elements[1].Tag.Group != 0x0043 || elements[1].Tag.Element != 0x0010 {
		t.Errorf("second element should be (0043,0010), got %v", elements[1].Tag)
	}

	// Verify software version tag
	if elements[2].Tag.Group != 0x0009 || elements[2].Tag.Element != 0x10E3 {
		t.Errorf("third element should be (0009,10E3), got %v", elements[2].Tag)
	}

	// Verify diffusion params tag
	if elements[3].Tag.Group != 0x0043 || elements[3].Tag.Element != 0x1039 {
		t.Errorf("fourth element should be (0043,1039), got %v", elements[3].Tag)
	}
}
