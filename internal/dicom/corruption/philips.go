package corruption

import (
	"fmt"
	"math/rand/v2"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

// generatePhilipsPrivateElements generates Philips private tags and sequences.
func generatePhilipsPrivateElements(rng *rand.Rand) []*dicom.Element {
	// Build a nested private sequence item at (2005,100E)
	scaleSlope := fmt.Sprintf("%.10f", rng.Float64()*100+1.0)
	scaleIntercept := fmt.Sprintf("%.10f", rng.Float64()*10-5.0)

	item := []*dicom.Element{
		mustNewPrivateElement(tag.Tag{Group: 0x2005, Element: 0x0011}, "LO", []string{"Philips MR Imaging DD 005"}),
		mustNewPrivateElement(tag.Tag{Group: 0x2005, Element: 0x1100}, "DS", []string{scaleSlope}),
		mustNewPrivateElement(tag.Tag{Group: 0x2005, Element: 0x1101}, "DS", []string{scaleIntercept}),
	}

	return []*dicom.Element{
		// Private creator blocks
		mustNewPrivateElement(tag.Tag{Group: 0x2001, Element: 0x0010}, "LO", []string{"Philips Imaging DD 001"}),
		mustNewPrivateElement(tag.Tag{Group: 0x2005, Element: 0x0010}, "LO", []string{"Philips MR Imaging DD 001"}),
		// Private sequence
		mustNewPrivateElement(tag.Tag{Group: 0x2005, Element: 0x100E}, "SQ", [][]*dicom.Element{item}),
	}
}
