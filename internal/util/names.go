package util

import (
	"math/rand/v2"
	"time"
)

// Package-level default RNG to avoid allocations when rng is nil
var defaultRNG = rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0))

// FrenchNameProbability is the probability (0.0-1.0) of generating a French name
const FrenchNameProbability = 0.20

var (
	// EnglishMaleFirstNames is the list of English male first names
	EnglishMaleFirstNames = []string{
		"James", "John", "Robert", "Michael", "William", "David", "Richard", "Joseph",
		"Thomas", "Charles", "Christopher", "Daniel", "Matthew", "Anthony", "Mark",
		"Donald", "Steven", "Paul", "Andrew", "Joshua", "Kenneth", "Kevin", "Brian",
		"George", "Timothy", "Ronald", "Edward", "Jason", "Jeffrey", "Ryan",
		"Jacob", "Gary", "Nicholas", "Eric", "Jonathan", "Stephen", "Larry", "Justin",
		"Scott", "Brandon", "Benjamin", "Samuel", "Raymond", "Gregory", "Frank", "Alexander",
		"Patrick", "Jack", "Dennis", "Jerry", "Tyler", "Aaron", "Jose", "Adam",
		"Nathan", "Henry", "Douglas", "Zachary", "Peter", "Kyle", "Noah", "Ethan",
		"Jeremy", "Walter", "Christian", "Keith", "Roger", "Terry", "Austin", "Sean",
		"Gerald", "Carl", "Dylan", "Harold", "Jordan", "Jesse", "Bryan", "Lawrence",
		"Arthur", "Gabriel", "Bruce", "Albert", "Willie", "Alan", "Wayne", "Billy",
		"Ralph", "Eugene", "Russell", "Bobby", "Mason", "Philip", "Louis", "Harry",
		"Vincent", "Logan", "Luke", "Caleb", "Evan", "Ian", "Connor", "Adrian",
		"Cole", "Dominic", "Elijah", "Gavin", "Isaac", "Jayden", "Landon", "Owen",
	}

	// EnglishFemaleFirstNames is the list of English female first names
	EnglishFemaleFirstNames = []string{
		"Mary", "Patricia", "Jennifer", "Linda", "Barbara", "Elizabeth", "Susan", "Jessica",
		"Sarah", "Karen", "Lisa", "Nancy", "Betty", "Margaret", "Sandra", "Ashley",
		"Kimberly", "Emily", "Donna", "Michelle", "Dorothy", "Carol", "Amanda", "Melissa",
		"Deborah", "Stephanie", "Rebecca", "Sharon", "Laura", "Cynthia", "Kathleen", "Amy",
		"Angela", "Shirley", "Anna", "Brenda", "Pamela", "Emma", "Nicole", "Helen",
		"Samantha", "Katherine", "Christine", "Debra", "Rachel", "Carolyn", "Janet", "Catherine",
		"Maria", "Heather", "Diane", "Ruth", "Julie", "Olivia", "Joyce", "Virginia",
		"Victoria", "Kelly", "Lauren", "Christina", "Joan", "Evelyn", "Judith", "Megan",
		"Andrea", "Cheryl", "Hannah", "Jacqueline", "Martha", "Gloria", "Teresa", "Ann",
		"Sara", "Madison", "Frances", "Kathryn", "Janice", "Jean", "Abigail", "Alice",
		"Julia", "Judy", "Sophia", "Grace", "Denise", "Amber", "Doris", "Marilyn",
		"Danielle", "Beverly", "Isabella", "Theresa", "Diana", "Natalie", "Brittany", "Charlotte",
		"Marie", "Kayla", "Alexis", "Lori", "Chloe", "Ava", "Mia", "Ella",
		"Lily", "Zoe", "Audrey", "Hazel", "Violet", "Aurora", "Savannah", "Brooklyn",
	}

	// EnglishLastNames is the list of English last names
	EnglishLastNames = []string{
		"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis",
		"Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson", "Thomas",
		"Taylor", "Moore", "Jackson", "Martin", "Lee", "Perez", "Thompson", "White",
		"Harris", "Sanchez", "Clark", "Ramirez", "Lewis", "Robinson", "Walker", "Young",
		"Allen", "King", "Wright", "Scott", "Torres", "Nguyen", "Hill", "Flores",
		"Green", "Adams", "Nelson", "Baker", "Hall", "Rivera", "Campbell", "Mitchell",
		"Carter", "Roberts", "Gomez", "Phillips", "Evans", "Turner", "Diaz", "Parker",
		"Cruz", "Edwards", "Collins", "Reyes", "Stewart", "Morris", "Morales", "Murphy",
		"Cook", "Rogers", "Gutierrez", "Ortiz", "Morgan", "Cooper", "Peterson", "Bailey",
		"Reed", "Kelly", "Howard", "Ramos", "Kim", "Cox", "Ward", "Richardson",
		"Watson", "Brooks", "Chavez", "Wood", "James", "Bennett", "Gray", "Mendoza",
		"Ruiz", "Hughes", "Price", "Alvarez", "Castillo", "Sanders", "Patel", "Myers",
		"Long", "Ross", "Foster", "Jimenez", "Powell", "Jenkins", "Perry", "Russell",
		"Sullivan", "Bell", "Coleman", "Butler", "Henderson", "Barnes", "Gonzales", "Fisher",
		"Vasquez", "Simmons", "Graham", "Mccoy", "Reynolds", "Hamilton", "Griffin", "Wallace",
		"West", "Cole", "Hayes", "Bryant", "Herrera", "Gibson", "Ellis", "Tran",
	}

	// FrenchMaleFirstNames is the list of French male first names
	FrenchMaleFirstNames = []string{
		"Jean", "Pierre", "Michel", "André", "Philippe", "Alain", "Bernard", "Jacques",
		"François", "Christian", "Daniel", "Patrick", "Nicolas", "Olivier", "Laurent",
		"Thierry", "Stéphane", "Éric", "David", "Julien", "Christophe", "Pascal",
		"Sébastien", "Marc", "Vincent", "Antoine", "Alexandre", "Maxime", "Thomas",
		"Lucas", "Hugo", "Louis", "Arthur", "Gabriel", "Raphaël", "Paul", "Jules",
		"Mathieu", "Romain", "Guillaume", "Benoît", "Cédric", "Fabien", "Yannick", "Hervé",
		"Didier", "Gilles", "Bruno", "Claude", "Serge", "Dominique", "Frédéric", "Emmanuel",
		"Arnaud", "Rémi", "Damien", "Adrien", "Florian", "Quentin", "Jérôme", "Xavier",
	}

	// FrenchFemaleFirstNames is the list of French female first names
	FrenchFemaleFirstNames = []string{
		"Marie", "Nathalie", "Isabelle", "Sylvie", "Catherine", "Françoise", "Valérie",
		"Christine", "Monique", "Sophie", "Patricia", "Martine", "Nicole", "Sandrine",
		"Stéphanie", "Céline", "Julie", "Aurélie", "Caroline", "Laurence", "Émilie",
		"Claire", "Anne", "Camille", "Laura", "Sarah", "Manon", "Emma", "Léa",
		"Chloé", "Zoé", "Alice", "Charlotte", "Lucie", "Juliette", "Louise",
		"Hélène", "Delphine", "Brigitte", "Véronique", "Corinne", "Annick", "Mireille", "Odile",
		"Élise", "Margaux", "Pauline", "Marine", "Morgane", "Anaïs", "Océane", "Inès",
		"Élodie", "Mathilde", "Clémence", "Justine", "Laure", "Agathe", "Estelle", "Noémie",
	}

	// FrenchLastNames is the list of French last names
	FrenchLastNames = []string{
		"Martin", "Bernard", "Dubois", "Thomas", "Robert", "Richard", "Petit",
		"Durand", "Leroy", "Moreau", "Simon", "Laurent", "Lefebvre", "Michel",
		"Garcia", "David", "Bertrand", "Roux", "Vincent", "Fournier", "Morel",
		"Girard", "André", "Lefevre", "Mercier", "Dupont", "Lambert", "Bonnet",
		"François", "Martinez", "Legrand", "Garnier", "Faure", "Rousseau", "Blanc",
		"Guerin", "Muller", "Henry", "Roussel", "Nicolas", "Perrin", "Morin",
		"Mathieu", "Clement", "Gauthier", "Dumont", "Lopez", "Fontaine", "Chevalier",
		"Robin", "Masson", "Sanchez", "Gerard", "Nguyen", "Boyer", "Denis", "Lemaire",
		"Dufour", "Renaud", "Barbier", "Arnaud", "Marchand", "Picard", "Leclerc", "Giraud",
		"Brun", "Gaillard", "Renard", "Roy", "Noel", "Meyer", "Hubert", "Gautier",
	}

	// MaleFirstNames combines English and French names for backward compatibility
	MaleFirstNames = append(EnglishMaleFirstNames, FrenchMaleFirstNames...)

	// FemaleFirstNames combines English and French names for backward compatibility
	FemaleFirstNames = append(EnglishFemaleFirstNames, FrenchFemaleFirstNames...)

	// LastNames combines English and French names for backward compatibility
	LastNames = append(EnglishLastNames, FrenchLastNames...)
)

// GeneratePatientName generates a realistic patient name based on sex.
// Names are 80% English and 20% French.
//
// Sex should be "M" or "F". Invalid values default to "F".
// If rng is nil, uses shared default RNG.
// Returns name in DICOM format: "LASTNAME^FIRSTNAME"
func GeneratePatientName(sex string, rng *rand.Rand) string {
	if rng == nil {
		rng = defaultRNG
	}

	// 20% chance of French name
	useFrench := rng.Float64() < FrenchNameProbability

	var firstName string
	var lastName string

	if useFrench {
		if sex == "M" {
			firstName = FrenchMaleFirstNames[rng.IntN(len(FrenchMaleFirstNames))]
		} else {
			firstName = FrenchFemaleFirstNames[rng.IntN(len(FrenchFemaleFirstNames))]
		}
		lastName = FrenchLastNames[rng.IntN(len(FrenchLastNames))]
	} else {
		if sex == "M" {
			firstName = EnglishMaleFirstNames[rng.IntN(len(EnglishMaleFirstNames))]
		} else {
			firstName = EnglishFemaleFirstNames[rng.IntN(len(EnglishFemaleFirstNames))]
		}
		lastName = EnglishLastNames[rng.IntN(len(EnglishLastNames))]
	}

	// DICOM format: LASTNAME^FIRSTNAME
	return lastName + "^" + firstName
}
