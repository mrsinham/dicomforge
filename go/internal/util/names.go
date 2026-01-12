package util

import (
	"math/rand/v2"
	"time"
)

// Package-level default RNG to avoid allocations when rng is nil
var defaultRNG = rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0))

var (
	// MaleFirstNames is the list of French male first names used for test data generation
	MaleFirstNames = []string{
		"Jean", "Pierre", "Michel", "André", "Philippe", "Alain", "Bernard", "Jacques",
		"François", "Christian", "Daniel", "Patrick", "Nicolas", "Olivier", "Laurent",
		"Thierry", "Stéphane", "Éric", "David", "Julien", "Christophe", "Pascal",
		"Sébastien", "Marc", "Vincent", "Antoine", "Alexandre", "Maxime", "Thomas",
		"Lucas", "Hugo", "Louis", "Arthur", "Gabriel", "Raphaël", "Paul", "Jules",
	}

	// FemaleFirstNames is the list of French female first names used for test data generation
	FemaleFirstNames = []string{
		"Marie", "Nathalie", "Isabelle", "Sylvie", "Catherine", "Françoise", "Valérie",
		"Christine", "Monique", "Sophie", "Patricia", "Martine", "Nicole", "Sandrine",
		"Stéphanie", "Céline", "Julie", "Aurélie", "Caroline", "Laurence", "Émilie",
		"Claire", "Anne", "Camille", "Laura", "Sarah", "Manon", "Emma", "Léa",
		"Chloé", "Zoé", "Alice", "Charlotte", "Lucie", "Juliette", "Louise",
	}

	// LastNames is the list of French last names used for test data generation
	LastNames = []string{
		"Martin", "Bernard", "Dubois", "Thomas", "Robert", "Richard", "Petit",
		"Durand", "Leroy", "Moreau", "Simon", "Laurent", "Lefebvre", "Michel",
		"Garcia", "David", "Bertrand", "Roux", "Vincent", "Fournier", "Morel",
		"Girard", "André", "Lefevre", "Mercier", "Dupont", "Lambert", "Bonnet",
		"François", "Martinez", "Legrand", "Garnier", "Faure", "Rousseau", "Blanc",
		"Guerin", "Muller", "Henry", "Roussel", "Nicolas", "Perrin", "Morin",
		"Mathieu", "Clement", "Gauthier", "Dumont", "Lopez", "Fontaine", "Chevalier",
		"Robin", "Masson", "Sanchez", "Gerard", "Nguyen", "Boyer", "Denis", "Lemaire",
	}
)

// GeneratePatientName generates a realistic French patient name based on sex.
//
// Sex should be "M" or "F". Invalid values default to "F".
// If rng is nil, uses shared default RNG.
// Returns name in DICOM format: "LASTNAME^FIRSTNAME"
func GeneratePatientName(sex string, rng *rand.Rand) string {
	if rng == nil {
		rng = defaultRNG
	}

	var firstName string
	if sex == "M" {
		firstName = MaleFirstNames[rng.IntN(len(MaleFirstNames))]
	} else {
		firstName = FemaleFirstNames[rng.IntN(len(FemaleFirstNames))]
	}

	lastName := LastNames[rng.IntN(len(LastNames))]

	// DICOM format: LASTNAME^FIRSTNAME
	return lastName + "^" + firstName
}
