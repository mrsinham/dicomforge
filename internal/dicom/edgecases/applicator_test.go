package edgecases

import (
	"math/rand/v2"
	"testing"
)

func TestApplicator_ShouldApply(t *testing.T) {
	config := Config{Percentage: 50, Types: []EdgeCaseType{SpecialChars}}
	rng := rand.New(rand.NewPCG(42, 42))
	app := NewApplicator(config, rng)

	// Test over 100 iterations
	applied := 0
	for i := 0; i < 100; i++ {
		if app.ShouldApply() {
			applied++
		}
	}
	// Should be roughly 50% (allow 30-70 range for randomness)
	if applied < 30 || applied > 70 {
		t.Errorf("50%% should apply ~50 times in 100, got %d", applied)
	}
}

func TestApplicator_SelectEdgeCaseType(t *testing.T) {
	config := Config{
		Percentage: 100,
		Types:      []EdgeCaseType{SpecialChars, LongNames},
	}
	rng := rand.New(rand.NewPCG(42, 42))
	app := NewApplicator(config, rng)

	selected := app.SelectEdgeCaseType()
	if selected != SpecialChars && selected != LongNames {
		t.Errorf("Selected type should be one of configured types: %v", selected)
	}
}

func TestApplicator_ApplyToPatientName(t *testing.T) {
	config := Config{Percentage: 100, Types: []EdgeCaseType{SpecialChars}}
	rng := rand.New(rand.NewPCG(42, 42))
	app := NewApplicator(config, rng)

	name := app.ApplyToPatientName("M", "SMITH^JOHN")
	// With special-chars, should get a special character name
	if name == "SMITH^JOHN" {
		t.Error("Edge case should modify the name")
	}
}

func TestApplicator_ApplyToPatientID(t *testing.T) {
	config := Config{Percentage: 100, Types: []EdgeCaseType{VariedIDs}}
	rng := rand.New(rand.NewPCG(42, 42))
	app := NewApplicator(config, rng)

	id := app.ApplyToPatientID("PID123456")
	if id == "PID123456" {
		t.Error("Edge case should modify the ID")
	}
}

func TestApplicator_GetTagsToOmit(t *testing.T) {
	config := Config{Percentage: 100, Types: []EdgeCaseType{MissingTags}}
	rng := rand.New(rand.NewPCG(42, 42))
	app := NewApplicator(config, rng)

	tags := app.GetTagsToOmit()
	if len(tags) == 0 {
		t.Error("Should return tags to omit when MissingTags is enabled")
	}
}

func TestApplicator_GetTagsToOmit_NotEnabled(t *testing.T) {
	config := Config{Percentage: 100, Types: []EdgeCaseType{SpecialChars}}
	rng := rand.New(rand.NewPCG(42, 42))
	app := NewApplicator(config, rng)

	tags := app.GetTagsToOmit()
	if len(tags) != 0 {
		t.Error("Should return empty when MissingTags is not enabled")
	}
}
