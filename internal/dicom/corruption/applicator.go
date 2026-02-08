package corruption

import (
	"fmt"
	"math/rand/v2"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

// mustNewPrivateElement creates a DICOM element with a private tag and explicit VR.
// This is required because dicom.NewElement fails on unregistered private tags.
func mustNewPrivateElement(t tag.Tag, rawVR string, data any) *dicom.Element {
	value, err := dicom.NewValue(data)
	if err != nil {
		panic(fmt.Sprintf("failed to create value for private element %v: %v", t, err))
	}
	return &dicom.Element{
		Tag:                    t,
		ValueRepresentation:    tag.GetVRKind(t, rawVR),
		RawValueRepresentation: rawVR,
		Value:                  value,
	}
}

// Applicator generates corruption elements based on the configured types.
type Applicator struct {
	config Config
	rng    *rand.Rand
}

// NewApplicator creates a new corruption applicator.
func NewApplicator(config Config, rng *rand.Rand) *Applicator {
	return &Applicator{config: config, rng: rng}
}

// GenerateCorruptionElements generates all corruption elements for the enabled types.
func (a *Applicator) GenerateCorruptionElements() []*dicom.Element {
	var elements []*dicom.Element

	if a.config.HasType(SiemensCSA) {
		elements = append(elements, generateSiemensCSAElements(a.rng)...)
	}
	if a.config.HasType(GEPrivate) {
		elements = append(elements, generateGEPrivateElements(a.rng)...)
	}
	if a.config.HasType(PhilipsPrivate) {
		elements = append(elements, generatePhilipsPrivateElements(a.rng)...)
	}
	if a.config.HasType(MalformedLengths) {
		elements = append(elements, generateMalformedPlaceholders()...)
	}

	return elements
}

// HasMalformedLengths returns true if malformed-lengths corruption is enabled.
func (a *Applicator) HasMalformedLengths() bool {
	return a.config.HasType(MalformedLengths)
}
