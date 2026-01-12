package util

import "math/rand/v2"

var (
	maleFirstNames = []string{
		"Jean", "Pierre", "Michel", "André", "Philippe", "Alain", "Bernard", "Jacques",
		"François", "Christian", "Daniel", "Patrick", "Nicolas", "Olivier", "Laurent",
		"Thierry", "Stéphane", "Éric", "David", "Julien", "Christophe", "Pascal",
		"Sébastien", "Marc", "Vincent", "Antoine", "Alexandre", "Maxime", "Thomas",
		"Lucas", "Hugo", "Louis", "Arthur", "Gabriel", "Raphaël", "Paul", "Jules",
	}

	femaleFirstNames = []string{
		"Marie", "Nathalie", "Isabelle", "Sylvie", "Catherine", "Françoise", "Valérie",
		"Christine", "Monique", "Sophie", "Patricia", "Martine", "Nicole", "Sandrine",
		"Stéphanie", "Céline", "Julie", "Aurélie", "Caroline", "Laurence", "Émilie",
		"Claire", "Anne", "Camille", "Laura", "Sarah", "Manon", "Emma", "Léa",
		"Chloé", "Zoé", "Alice", "Charlotte", "Lucie", "Juliette", "Louise",
	}

	lastNames = []string{
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
// Sex should be "M" or "F". Returns name in DICOM format: "LASTNAME^FIRSTNAME"
func GeneratePatientName(sex string) string {
	var firstName string
	if sex == "M" {
		firstName = maleFirstNames[rand.IntN(len(maleFirstNames))]
	} else {
		firstName = femaleFirstNames[rand.IntN(len(femaleFirstNames))]
	}

	lastName := lastNames[rand.IntN(len(lastNames))]

	// DICOM format: LASTNAME^FIRSTNAME
	return lastName + "^" + firstName
}
