package modalities

import (
	"fmt"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

// mustNewElement creates a new DICOM element, panicking on error.
func mustNewElement(t tag.Tag, value interface{}) *dicom.Element {
	elem, err := dicom.NewElement(t, value)
	if err != nil {
		panic(fmt.Sprintf("failed to create element %v: %v", t, err))
	}
	return elem
}

// floatToDS converts a float64 to a DICOM Decimal String.
func floatToDS(f float64) string {
	return fmt.Sprintf("%.6g", f)
}

// intToIS converts an int to a DICOM Integer String.
func intToIS(i int) string {
	return fmt.Sprintf("%d", i)
}
