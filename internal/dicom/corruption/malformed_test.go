package corruption

import (
	"encoding/binary"
	"testing"
)

func TestGenerateMalformedPlaceholders(t *testing.T) {
	elements := generateMalformedPlaceholders()

	if len(elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elements))
	}

	// Verify FL placeholder (written as private OB at 0071,0010)
	if elements[0].Tag.Group != 0x0071 || elements[0].Tag.Element != 0x0010 {
		t.Errorf("placeholder should be (0071,0010), got %v", elements[0].Tag)
	}
	if elements[0].RawValueRepresentation != "OB" {
		t.Errorf("placeholder should have OB VR, got %s", elements[0].RawValueRepresentation)
	}
}

func TestRewriteTagAndPatch(t *testing.T) {
	// Build a data segment with placeholder (0071,0010) OB long-form
	// Layout: Group(2) | Element(2) | VR(2) | Reserved(2) | VL(4) | Data(8)
	data := []byte{
		0x71, 0x00, // Group 0x0071 (LE)
		0x10, 0x00, // Element 0x0010 (LE)
		'O', 'B', // VR = "OB"
		0x00, 0x00, // Reserved
		0x08, 0x00, 0x00, 0x00, // VL = 8
		0x00, 0x00, 0x80, 0x3F, // 1.0f
		0x00, 0x00, 0x00, 0x40, // 2.0f
	}

	ok := rewriteTagAndPatch(data, 0x0071, 0x0010, 0x0070, 0x0253, "FL", 7)
	if !ok {
		t.Fatal("expected rewriteTagAndPatch to return true")
	}

	// Verify tag was rewritten to (0070,0253)
	group := binary.LittleEndian.Uint16(data[0:2])
	elem := binary.LittleEndian.Uint16(data[2:4])
	if group != 0x0070 || elem != 0x0253 {
		t.Errorf("tag should be (0070,0253), got (%04X,%04X)", group, elem)
	}

	// Verify VR was rewritten to "FL"
	vr := string(data[4:6])
	if vr != "FL" {
		t.Errorf("VR should be FL, got %s", vr)
	}

	// FL is short-form: VL at offset 6-8
	vl := binary.LittleEndian.Uint16(data[6:8])
	if vl != 7 {
		t.Errorf("VL should be 7, got %d", vl)
	}
}

func TestPatchPixelDataOddLength(t *testing.T) {
	// Build a PixelData element (7FE0,0010) OW long-form
	// Layout: Group(2) | Element(2) | VR(2) | Reserved(2) | VL(4)
	data := []byte{
		0xE0, 0x7F, // Group 0x7FE0 (LE)
		0x10, 0x00, // Element 0x0010 (LE)
		'O', 'W', // VR = "OW"
		0x00, 0x00, // Reserved
		0x00, 0x00, 0x02, 0x00, // VL = 131072 (even)
		// (pixel data would follow)
	}

	ok := patchPixelDataOddLength(data)
	if !ok {
		t.Fatal("expected patchPixelDataOddLength to return true")
	}

	vl := binary.LittleEndian.Uint32(data[8:12])
	if vl != 131071 {
		t.Errorf("VL should be 131071 (odd), got %d", vl)
	}
	if vl%2 == 0 {
		t.Errorf("VL should be odd, got %d", vl)
	}
}

func TestPatchPixelDataOddLength_AlreadyOdd(t *testing.T) {
	data := []byte{
		0xE0, 0x7F,
		0x10, 0x00,
		'O', 'W',
		0x00, 0x00,
		0x07, 0x00, 0x00, 0x00, // VL = 7 (already odd)
	}

	ok := patchPixelDataOddLength(data)
	if ok {
		t.Error("should not patch already-odd VL")
	}
}

func TestPatchPixelDataOddLength_NotFound(t *testing.T) {
	data := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	ok := patchPixelDataOddLength(data)
	if ok {
		t.Error("should return false when PixelData not found")
	}
}

func TestRewriteTagAndPatch_NotFound(t *testing.T) {
	data := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	ok := rewriteTagAndPatch(data, 0x0071, 0x0010, 0x0070, 0x0253, "FL", 7)
	if ok {
		t.Error("should return false when tag not found")
	}
}
