package edgecases

import (
	"math/rand/v2"
	"strconv"
	"testing"
	"time"
)

func TestGenerateOldBirthDate(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))
	date := GenerateOldBirthDate(rng)
	if len(date) != 8 {
		t.Errorf("Date should be YYYYMMDD format, got %s", date)
	}
	year, _ := strconv.Atoi(date[:4])
	if year > 1950 {
		t.Errorf("Old birth date should be <= 1950, got %d", year)
	}
}

func TestGeneratePartialDate(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))
	for i := 0; i < 10; i++ {
		date := GeneratePartialDate(rng)
		if len(date) != 4 && len(date) != 6 {
			t.Errorf("Partial date should be YYYY or YYYYMM, got %s (len=%d)", date, len(date))
		}
	}
}

func TestGenerateFutureStudyDate(t *testing.T) {
	rng := rand.New(rand.NewPCG(42, 42))
	date := GenerateFutureStudyDate(rng)
	if len(date) != 8 {
		t.Errorf("Date should be YYYYMMDD format, got %s", date)
	}
	year, _ := strconv.Atoi(date[:4])
	if year <= time.Now().Year() {
		t.Errorf("Future date should be > current year, got %d", year)
	}
}
