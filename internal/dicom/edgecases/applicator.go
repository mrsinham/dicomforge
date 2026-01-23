package edgecases

import "math/rand/v2"

// Applicator applies edge cases to generated values
type Applicator struct {
	config Config
	rng    *rand.Rand
}

// NewApplicator creates a new edge case applicator
func NewApplicator(config Config, rng *rand.Rand) *Applicator {
	return &Applicator{config: config, rng: rng}
}

// ShouldApply returns true if edge cases should apply to this file
func (a *Applicator) ShouldApply() bool {
	return a.rng.IntN(100) < a.config.Percentage
}

// SelectEdgeCaseType randomly selects which edge case type to apply
func (a *Applicator) SelectEdgeCaseType() EdgeCaseType {
	return a.config.Types[a.rng.IntN(len(a.config.Types))]
}

// ApplyToPatientName applies edge cases to a patient name
func (a *Applicator) ApplyToPatientName(sex, original string) string {
	edgeType := a.SelectEdgeCaseType()
	switch edgeType {
	case SpecialChars:
		return GenerateSpecialCharName(sex, a.rng)
	case LongNames:
		return GenerateLongPatientName(sex, a.rng)
	default:
		return original
	}
}

// ApplyToPatientID applies edge cases to a patient ID
func (a *Applicator) ApplyToPatientID(original string) string {
	edgeType := a.SelectEdgeCaseType()
	switch edgeType {
	case VariedIDs:
		return GenerateRandomVariedPatientID(a.rng)
	case LongNames:
		return GenerateLongPatientID(a.rng)
	default:
		return original
	}
}

// ApplyToBirthDate applies edge cases to a birth date
func (a *Applicator) ApplyToBirthDate(original string) string {
	edgeType := a.SelectEdgeCaseType()
	switch edgeType {
	case OldDates:
		if a.rng.IntN(2) == 0 {
			return GenerateOldBirthDate(a.rng)
		}
		return GeneratePartialDate(a.rng)
	default:
		return original
	}
}

// ApplyToStudyDate applies edge cases to a study date
func (a *Applicator) ApplyToStudyDate(original string) string {
	if a.config.HasType(OldDates) && a.rng.IntN(4) == 0 {
		return GenerateFutureStudyDate(a.rng)
	}
	return original
}

// GetTagsToOmit returns tags that should be omitted for this file
func (a *Applicator) GetTagsToOmit() []string {
	if !a.config.HasType(MissingTags) {
		return nil
	}
	count := 1 + a.rng.IntN(3) // Omit 1-3 tags
	return SelectTagsToOmit(a.rng, count)
}
