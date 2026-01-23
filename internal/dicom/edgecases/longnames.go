package edgecases

import (
	"math/rand/v2"
	"strings"
)

const DICOMLOMaxLength = 64

var longLastNames = []string{
	"ALEXANDROPOULOSWILLIAMSONBERG",
	"VANDENBERGHEMONTGOMERYSMITH",
	"CHRISTODOULOPOULOSSMITHBAUER",
	"SCHWARZENEGGERBAUERWILLIAMS",
	"MCCARTHYWILKINSONTHOMPSON",
}

var longFirstNames = []string{
	"ALEXANDERMAXIMILIANWILLIAM",
	"CHRISTOPHERJOHNATHANMICHAEL",
	"ELIZABETHCATHERINEANNAMARIE",
	"MARGARETISABELLAVICTORIAJANE",
	"BENJAMINFREDERICKNATHANJOHN",
}

// GenerateLongPatientName generates a patient name close to max DICOM length
func GenerateLongPatientName(sex string, rng *rand.Rand) string {
	lastName := longLastNames[rng.IntN(len(longLastNames))]
	firstName := longFirstNames[rng.IntN(len(longFirstNames))]
	name := lastName + "^" + firstName
	if len(name) > DICOMLOMaxLength {
		name = name[:DICOMLOMaxLength]
	}
	return name
}

// GenerateLongPatientID generates a PatientID at max length
func GenerateLongPatientID(rng *rand.Rand) string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var sb strings.Builder
	for i := 0; i < DICOMLOMaxLength; i++ {
		sb.WriteByte(chars[rng.IntN(len(chars))])
	}
	return sb.String()
}

// GenerateLongStudyDescription generates a StudyDescription at max length
func GenerateLongStudyDescription(rng *rand.Rand) string {
	descriptions := []string{
		"MRI BRAIN WITH AND WITHOUT CONTRAST DETAILED EXAMINATION FOR SUSPECTED LESION",
		"CT ABDOMEN PELVIS WITH CONTRAST COMPREHENSIVE EVALUATION FOLLOW UP EXAMINATION",
		"PET CT WHOLE BODY SCAN WITH FDG FOR ONCOLOGIC STAGING AND RESTAGING PURPOSES",
	}
	desc := descriptions[rng.IntN(len(descriptions))]
	if len(desc) > DICOMLOMaxLength {
		desc = desc[:DICOMLOMaxLength]
	}
	return desc
}
