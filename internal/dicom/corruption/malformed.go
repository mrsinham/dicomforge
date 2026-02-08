package corruption

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

// Malformed target tags - we place valid elements at these locations during generation,
// then patch their value lengths to invalid values in post-processing.
var (
	// (0071,0010) - A private tag with FL VR
	// We'll patch it to have a length not divisible by 4.
	malformedFLTag = tag.Tag{Group: 0x0071, Element: 0x0010}

	// (0069,0010) - A private tag with OW VR
	// We'll patch it to have an odd byte count.
	malformedOWTag = tag.Tag{Group: 0x0069, Element: 0x0010}
)

// generateMalformedPlaceholders creates placeholder elements at the target tags.
// These are valid elements that will be patched with incorrect lengths after writing.
func generateMalformedPlaceholders() []*dicom.Element {
	// FL element: 2 float32 values = 8 bytes (valid), will be patched to 7
	flData := []float64{1.0, 2.0}

	// OW element: 8 bytes (valid), will be patched to 7 (odd)
	owData := []byte{0x01, 0x00, 0x02, 0x00, 0x03, 0x00, 0x04, 0x00}

	return []*dicom.Element{
		mustNewPrivateElement(malformedOWTag, "OW", owData),
		mustNewPrivateElement(malformedFLTag, "FL", flData),
	}
}

// PatchMalformedLengths performs binary post-processing on a written DICOM file
// to patch value lengths to intentionally incorrect values.
func PatchMalformedLengths(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file for malformed patching: %w", err)
	}

	patched := false

	// Patch FL element at (0071,0010): find the tag bytes and patch VL to 7
	patched = patchTagValueLength(data, 0x0071, 0x0010, 7) || patched

	// Patch OW element at (0069,0010): find the tag bytes and patch VL to 7 (odd)
	patched = patchTagValueLength(data, 0x0069, 0x0010, 7) || patched

	if !patched {
		return nil // Nothing to patch
	}

	return os.WriteFile(filePath, data, 0600)
}

// patchTagValueLength finds a DICOM tag in binary data and patches its value length.
// For Explicit VR Little Endian, the layout is:
//
//	Group(2) | Element(2) | VR(2) | VL(2 or 6)
//
// For VRs like OB, OW, OF, SQ, UC, UN, UR, UT, the layout is:
//
//	Group(2) | Element(2) | VR(2) | Reserved(2) | VL(4)
//
// For short VRs like FL, US, SS, etc.:
//
//	Group(2) | Element(2) | VR(2) | VL(2)
func patchTagValueLength(data []byte, group, element uint16, newLength uint32) bool {
	// Build the tag bytes (Little Endian)
	tagBytes := make([]byte, 4)
	binary.LittleEndian.PutUint16(tagBytes[0:2], group)
	binary.LittleEndian.PutUint16(tagBytes[2:4], element)

	for i := 0; i <= len(data)-8; i++ {
		if data[i] == tagBytes[0] && data[i+1] == tagBytes[1] &&
			data[i+2] == tagBytes[2] && data[i+3] == tagBytes[3] {
			// Found the tag - determine VR type
			vrStr := string(data[i+4 : i+6])

			switch vrStr {
			case "OB", "OW", "OF", "SQ", "UC", "UN", "UR", "UT":
				// Long form: VR(2) + Reserved(2) + VL(4)
				if i+12 <= len(data) {
					binary.LittleEndian.PutUint32(data[i+8:i+12], newLength)
					return true
				}
			default:
				// Short form: VR(2) + VL(2)
				if i+8 <= len(data) {
					binary.LittleEndian.PutUint16(data[i+6:i+8], uint16(newLength))
					return true
				}
			}
		}
	}
	return false
}
