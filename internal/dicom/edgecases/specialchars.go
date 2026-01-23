package edgecases

import "math/rand/v2"

var specialCharFirstNamesMale = []string{
	"Jean-Pierre", "François", "André", "José", "Ángel",
	"Søren", "Björn", "Łukasz", "Jürgen", "O'Brien",
}

var specialCharFirstNamesFemale = []string{
	"Marie-Claire", "Françoise", "Éléonore", "María", "Ángela",
	"Siân", "Zoë", "Renée", "Hélène", "O'Hara",
}

var specialCharLastNames = []string{
	"Müller-Schmidt", "O'Connor", "D'Agostino", "García-López",
	"Björnsson", "Østergaard", "Çelik", "Škvorecký",
	"González", "Pérez-Rodríguez",
}

// GenerateSpecialCharName generates a patient name with special characters
func GenerateSpecialCharName(sex string, rng *rand.Rand) string {
	var firstName string
	if sex == "F" {
		firstName = specialCharFirstNamesFemale[rng.IntN(len(specialCharFirstNamesFemale))]
	} else {
		firstName = specialCharFirstNamesMale[rng.IntN(len(specialCharFirstNamesMale))]
	}
	lastName := specialCharLastNames[rng.IntN(len(specialCharLastNames))]
	return lastName + "^" + firstName
}
