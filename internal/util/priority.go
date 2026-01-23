// internal/util/priority.go
package util

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

// Priority represents exam priority level
type Priority int

const (
	PriorityRoutine Priority = iota
	PriorityHigh
	PriorityLow
)

// String returns the DICOM string representation of the priority
func (p Priority) String() string {
	switch p {
	case PriorityHigh:
		return "HIGH"
	case PriorityLow:
		return "LOW"
	default:
		return "ROUTINE"
	}
}

// ParsePriority parses a string into a Priority
func ParsePriority(s string) (Priority, error) {
	switch strings.ToUpper(s) {
	case "HIGH":
		return PriorityHigh, nil
	case "ROUTINE":
		return PriorityRoutine, nil
	case "LOW":
		return PriorityLow, nil
	default:
		return PriorityRoutine, fmt.Errorf("invalid priority: %s (valid: HIGH, ROUTINE, LOW)", s)
	}
}

// GeneratePriority generates a random priority with realistic distribution.
// Distribution: 70% ROUTINE, 20% HIGH, 10% LOW
func GeneratePriority(rng *rand.Rand) Priority {
	if rng == nil {
		rng = defaultRNG
	}

	r := rng.Float64()
	if r < 0.70 {
		return PriorityRoutine
	} else if r < 0.90 {
		return PriorityHigh
	}
	return PriorityLow
}
