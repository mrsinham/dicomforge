package corruption

import (
	"math/rand/v2"
	"testing"
)

func TestBuildCSAHeader(t *testing.T) {
	elements := []csaElement{
		{
			Name: "TestElement", VM: 1, VR: "IS", SyngoDT: 6, NumItems: 1,
			Values: []string{"42"},
		},
	}

	data := buildCSAHeader(elements)

	// Verify magic bytes
	if string(data[0:4]) != "SV10" {
		t.Errorf("expected SV10 magic, got %q", string(data[0:4]))
	}
	if data[4] != 0x04 || data[5] != 0x03 || data[6] != 0x02 || data[7] != 0x01 {
		t.Error("incorrect secondary magic bytes")
	}
}

func TestGenerateCSAImageHeader(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))
	header := generateCSAImageHeader(rng)

	// Should start with SV10
	if len(header) < 8 {
		t.Fatal("header too short")
	}
	if string(header[0:4]) != "SV10" {
		t.Errorf("expected SV10 magic, got %q", string(header[0:4]))
	}
	// Should be in realistic size range (5-15KB)
	if len(header) < 1024 {
		t.Errorf("header too small: %d bytes", len(header))
	}
}

func TestGenerateCSASeriesHeader(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))
	header := generateCSASeriesHeader(rng)

	if len(header) < 8 {
		t.Fatal("header too short")
	}
	if string(header[0:4]) != "SV10" {
		t.Errorf("expected SV10 magic, got %q", string(header[0:4]))
	}
}

func TestGenerateSiemensCSAElements(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))
	elements := generateSiemensCSAElements(rng)

	if len(elements) != 4 {
		t.Fatalf("expected 4 elements, got %d", len(elements))
	}

	// Verify private creator
	if elements[0].Tag.Group != 0x0029 || elements[0].Tag.Element != 0x0010 {
		t.Errorf("first element should be private creator (0029,0010), got %v", elements[0].Tag)
	}

	// Verify CSA Image Header tag
	if elements[1].Tag.Group != 0x0029 || elements[1].Tag.Element != 0x1010 {
		t.Errorf("second element should be CSA Image Header (0029,1010), got %v", elements[1].Tag)
	}

	// Verify CSA Series Header tag
	if elements[2].Tag.Group != 0x0029 || elements[2].Tag.Element != 0x1020 {
		t.Errorf("third element should be CSA Series Header (0029,1020), got %v", elements[2].Tag)
	}

	// Verify crash-trigger SQ tag
	if elements[3].Tag.Group != 0x0029 || elements[3].Tag.Element != 0x1102 {
		t.Errorf("fourth element should be crash-trigger SQ (0029,1102), got %v", elements[3].Tag)
	}
}
