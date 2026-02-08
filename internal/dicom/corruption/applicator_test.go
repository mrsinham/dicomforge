package corruption

import (
	"math/rand/v2"
	"testing"
)

func TestApplicator_GenerateCorruptionElements(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))

	tests := []struct {
		name     string
		types    []CorruptionType
		minCount int // minimum expected elements
	}{
		{
			name:     "siemens only",
			types:    []CorruptionType{SiemensCSA},
			minCount: 4, // creator + image header + series header + SQ
		},
		{
			name:     "ge only",
			types:    []CorruptionType{GEPrivate},
			minCount: 4, // 2 creators + software version + diffusion params
		},
		{
			name:     "philips only",
			types:    []CorruptionType{PhilipsPrivate},
			minCount: 3, // 2 creators + sequence
		},
		{
			name:     "malformed only",
			types:    []CorruptionType{MalformedLengths},
			minCount: 1, // FL placeholder (PixelData patched in post-processing)
		},
		{
			name:     "all types",
			types:    AllCorruptionTypes(),
			minCount: 12, // 4 + 4 + 3 + 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{Types: tt.types}
			applicator := NewApplicator(config, rng)

			elements := applicator.GenerateCorruptionElements()
			if len(elements) < tt.minCount {
				t.Errorf("GenerateCorruptionElements() returned %d elements, want at least %d", len(elements), tt.minCount)
			}

			// Verify all elements have valid tags
			for _, elem := range elements {
				if elem.Tag.Group == 0 && elem.Tag.Element == 0 {
					t.Error("element has zero tag")
				}
			}
		})
	}
}

func TestApplicator_HasMalformedLengths(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))

	withMalformed := NewApplicator(Config{Types: []CorruptionType{MalformedLengths}}, rng)
	if !withMalformed.HasMalformedLengths() {
		t.Error("should have malformed lengths")
	}

	withoutMalformed := NewApplicator(Config{Types: []CorruptionType{SiemensCSA}}, rng)
	if withoutMalformed.HasMalformedLengths() {
		t.Error("should not have malformed lengths")
	}
}
