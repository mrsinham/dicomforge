package edgecases

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

// IDFormat represents different PatientID formats
type IDFormat int

const (
	IDWithDashes  IDFormat = iota // e.g., "123-456-789"
	IDWithLetters                 // e.g., "ABC123DEF"
	IDWithSpaces                  // e.g., "PAT 12345 67"
	IDLong                        // e.g., 64-character ID
	IDMixed                       // e.g., "PT-2024-ABC 123"
)

// GenerateVariedPatientID generates a PatientID in the specified format
func GenerateVariedPatientID(format IDFormat, rng *rand.Rand) string {
	switch format {
	case IDWithDashes:
		return fmt.Sprintf("%03d-%03d-%03d", rng.IntN(1000), rng.IntN(1000), rng.IntN(1000))
	case IDWithLetters:
		letters := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		var sb strings.Builder
		for i := 0; i < 10; i++ {
			if i%2 == 0 {
				sb.WriteByte(letters[rng.IntN(len(letters))])
			} else {
				sb.WriteByte('0' + byte(rng.IntN(10)))
			}
		}
		return sb.String()
	case IDWithSpaces:
		return fmt.Sprintf("PAT %05d %02d", rng.IntN(100000), rng.IntN(100))
	case IDLong:
		return GenerateLongPatientID(rng)
	case IDMixed:
		return fmt.Sprintf("PT-%04d-%c%c%c %03d",
			rng.IntN(10000),
			'A'+byte(rng.IntN(26)),
			'A'+byte(rng.IntN(26)),
			'A'+byte(rng.IntN(26)),
			rng.IntN(1000))
	default:
		return fmt.Sprintf("PAT%06d", rng.IntN(1000000))
	}
}

// GenerateRandomVariedPatientID randomly selects a format
func GenerateRandomVariedPatientID(rng *rand.Rand) string {
	format := IDFormat(rng.IntN(5))
	return GenerateVariedPatientID(format, rng)
}
