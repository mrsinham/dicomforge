package edgecases

import (
	"fmt"
	"math/rand/v2"
	"time"
)

// GenerateOldBirthDate generates a very old birth date (1900-1950)
func GenerateOldBirthDate(rng *rand.Rand) string {
	year := 1900 + rng.IntN(51) // 1900-1950
	month := 1 + rng.IntN(12)
	day := 1 + rng.IntN(28)
	return fmt.Sprintf("%04d%02d%02d", year, month, day)
}

// GeneratePartialDate generates a partial DICOM date (YYYY or YYYYMM)
func GeneratePartialDate(rng *rand.Rand) string {
	year := 1950 + rng.IntN(50) // 1950-1999
	if rng.IntN(2) == 0 {
		// Year only
		return fmt.Sprintf("%04d", year)
	}
	// Year and month
	month := 1 + rng.IntN(12)
	return fmt.Sprintf("%04d%02d", year, month)
}

// GenerateFutureStudyDate generates a study date in the future
func GenerateFutureStudyDate(rng *rand.Rand) string {
	year := time.Now().Year() + 1 + rng.IntN(5) // 1-5 years in future
	month := 1 + rng.IntN(12)
	day := 1 + rng.IntN(28)
	return fmt.Sprintf("%04d%02d%02d", year, month, day)
}
