package edgecases

import "math/rand/v2"

// OptionalTags lists DICOM tags that can be safely omitted
var OptionalTags = []string{
	"BodyPartExamined",
	"StudyDescription",
	"SeriesDescription",
	"InstitutionName",
	"ReferringPhysicianName",
	"PerformingPhysicianName",
	"OperatorsName",
	"ProtocolName",
}

// SelectTagsToOmit randomly selects which optional tags to omit
func SelectTagsToOmit(rng *rand.Rand, count int) []string {
	if count >= len(OptionalTags) {
		return OptionalTags
	}
	// Fisher-Yates shuffle and take first count
	indices := make([]int, len(OptionalTags))
	for i := range indices {
		indices[i] = i
	}
	for i := len(indices) - 1; i > 0; i-- {
		j := rng.IntN(i + 1)
		indices[i], indices[j] = indices[j], indices[i]
	}
	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = OptionalTags[indices[i]]
	}
	return result
}
