package corruption

import (
	"fmt"
	"math/rand/v2"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

// generateGEPrivateElements generates GE GEMS private tags.
func generateGEPrivateElements(rng *rand.Rand) []*dicom.Element {
	// Generate a realistic GE software version string
	major := rng.IntN(10) + 20
	minor := rng.IntN(10)
	patch := rng.IntN(100)
	softwareVersion := fmt.Sprintf("DV%d.%d_%d_M5", major, minor, patch)

	// Generate multi-valued diffusion parameters (4 values as per GE convention)
	diffusionValues := make([]string, 4)
	for i := range diffusionValues {
		diffusionValues[i] = fmt.Sprintf("%d", rng.IntN(1000))
	}

	return []*dicom.Element{
		// Private creator blocks
		mustNewPrivateElement(tag.Tag{Group: 0x0009, Element: 0x0010}, "LO", []string{"GEMS_IDEN_01"}),
		mustNewPrivateElement(tag.Tag{Group: 0x0043, Element: 0x0010}, "LO", []string{"GEMS_PARM_01"}),
		// GE software version
		mustNewPrivateElement(tag.Tag{Group: 0x0009, Element: 0x10E3}, "LO", []string{softwareVersion}),
		// GE diffusion parameters (multi-valued)
		mustNewPrivateElement(tag.Tag{Group: 0x0043, Element: 0x1039}, "IS", diffusionValues),
	}
}
