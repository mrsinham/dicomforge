package corruption

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

// These reproduce the exact malformed elements seen in real Siemens scanner output:
//
//	W: DcmItem: Length of element (0070,0253) is not a multiple of 4 (VR=FL)
//	W: DcmItem: Length of element (7fe0,0010) is not a multiple of 2 (VR=OW)
//	(0029,1102) SQ (Sequence with explicit length #=1)  # 9434, 1 Unknown Tag & Data
var (
	// (0070,0253) LineThickness - standard FL tag.
	// Real Siemens files have this with a value length not divisible by 4.
	malformedFLTag = tag.Tag{Group: 0x0070, Element: 0x0253}
)

// generateMalformedPlaceholders creates placeholder elements at the target tags.
// These are valid elements that will be patched with incorrect lengths after writing.
// The FL placeholder uses a private tag to avoid VR type-checking by the DICOM writer;
// PatchMalformedLengths then patches both this tag AND the PixelData (7FE0,0010) tag
// to reproduce the real dcmdump warnings.
func generateMalformedPlaceholders() []*dicom.Element {
	// FL element written as a private OB tag to bypass the library's VR type checks.
	// PatchMalformedLengths will rewrite the tag bytes to (0070,0253) with VR=FL
	// and a non-multiple-of-4 length, exactly as seen in real Siemens output.
	flPlaceholder := mustNewPrivateElement(
		tag.Tag{Group: 0x0071, Element: 0x0010}, "OB",
		[]byte{0x00, 0x00, 0x80, 0x3F, 0x00, 0x00, 0x00, 0x40}, // 1.0f, 2.0f as raw bytes
	)

	return []*dicom.Element{flPlaceholder}
}

// PatchMalformedLengths performs binary post-processing on a written DICOM file
// to reproduce the exact malformed elements from real Siemens scanner output.
//
// It patches:
//   - (0071,0010) OB placeholder -> rewritten to (0070,0253) FL with VL=7 (not multiple of 4)
//   - (7FE0,0010) PixelData OW -> VL patched to odd value (not multiple of 2)
func PatchMalformedLengths(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read file for malformed patching: %w", err)
	}

	patched := false

	// Rewrite the placeholder (0071,0010) OB -> (0070,0253) FL with VL=7
	patched = rewriteTagAndPatch(data, 0x0071, 0x0010, 0x0070, 0x0253, "FL", 7) || patched

	// Patch PixelData (7FE0,0010) OW -> odd VL (original VL minus 1)
	patched = patchPixelDataOddLength(data) || patched

	if !patched {
		return nil
	}

	return os.WriteFile(filePath, data, 0600)
}

// rewriteTagAndPatch finds an element by its original tag, rewrites it to a new tag
// with a new VR and patched value length. This is used to transform a placeholder
// private tag into the real standard tag with intentionally wrong VR length.
func rewriteTagAndPatch(data []byte, origGroup, origElem, newGroup, newElem uint16, newVR string, newVL uint32) bool {
	origTagBytes := make([]byte, 4)
	binary.LittleEndian.PutUint16(origTagBytes[0:2], origGroup)
	binary.LittleEndian.PutUint16(origTagBytes[2:4], origElem)

	for i := 0; i <= len(data)-12; i++ {
		if data[i] == origTagBytes[0] && data[i+1] == origTagBytes[1] &&
			data[i+2] == origTagBytes[2] && data[i+3] == origTagBytes[3] {

			// Rewrite group and element
			binary.LittleEndian.PutUint16(data[i:i+2], newGroup)
			binary.LittleEndian.PutUint16(data[i+2:i+4], newElem)

			// Rewrite VR
			copy(data[i+4:i+6], newVR)

			// Determine VL position based on new VR
			switch newVR {
			case "OB", "OW", "OF", "SQ", "UC", "UN", "UR", "UT":
				// Long form: VR(2) + Reserved(2) + VL(4)
				data[i+6] = 0x00
				data[i+7] = 0x00
				binary.LittleEndian.PutUint32(data[i+8:i+12], newVL)
			default:
				// Short form: VR(2) + VL(2)
				binary.LittleEndian.PutUint16(data[i+6:i+8], uint16(newVL))
			}
			return true
		}
	}
	return false
}

// patchPixelDataOddLength finds the PixelData element (7FE0,0010) and patches its
// value length to an odd number (original - 1), reproducing the dcmdump warning:
// "Length of element (7fe0,0010) is not a multiple of 2 (VR=OW)"
func patchPixelDataOddLength(data []byte) bool {
	// PixelData tag bytes: 0xE0, 0x7F, 0x10, 0x00 (Little Endian)
	for i := 0; i <= len(data)-12; i++ {
		if data[i] == 0xE0 && data[i+1] == 0x7F &&
			data[i+2] == 0x10 && data[i+3] == 0x00 {
			vrStr := string(data[i+4 : i+6])
			if vrStr == "OW" || vrStr == "OB" {
				// Long form: VR(2) + Reserved(2) + VL(4)
				currentVL := binary.LittleEndian.Uint32(data[i+8 : i+12])
				if currentVL > 1 && currentVL%2 == 0 {
					// Make it odd
					binary.LittleEndian.PutUint32(data[i+8:i+12], currentVL-1)
					return true
				}
			}
		}
	}
	return false
}
