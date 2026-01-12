package dicom

import (
	"testing"
)

func TestGenerateMetadata_BasicStructure(t *testing.T) {
	opts := MetadataOptions{
		NumImages:      10,
		Width:          256,
		Height:         256,
		InstanceNumber: 1,
		PatientID:      "TEST123",
		PatientName:    "DOE^JOHN",
		StudyUID:       "1.2.3.4.5",
		SeriesUID:      "1.2.3.4.6",
	}

	ds, err := GenerateMetadata(opts)
	if err != nil {
		t.Fatalf("GenerateMetadata failed: %v", err)
	}

	if ds == nil {
		t.Fatal("Expected non-nil dataset")
	}
}
