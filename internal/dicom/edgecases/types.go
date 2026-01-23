package edgecases

import (
	"fmt"
	"strings"
)

// EdgeCaseType represents a category of edge case
type EdgeCaseType string

const (
	SpecialChars EdgeCaseType = "special-chars"
	LongNames    EdgeCaseType = "long-names"
	MissingTags  EdgeCaseType = "missing-tags"
	OldDates     EdgeCaseType = "old-dates"
	VariedIDs    EdgeCaseType = "varied-ids"
)

// AllEdgeCaseTypes returns all valid edge case types
func AllEdgeCaseTypes() []EdgeCaseType {
	return []EdgeCaseType{SpecialChars, LongNames, MissingTags, OldDates, VariedIDs}
}

// Config holds edge case generation settings
type Config struct {
	Percentage int            // 0-100, percentage of files to apply edge cases
	Types      []EdgeCaseType // Which edge case types to enable
}

// ParseTypes parses comma-separated edge case types
func ParseTypes(input string) ([]EdgeCaseType, error) {
	if input == "" {
		return nil, nil
	}
	parts := strings.Split(input, ",")
	result := make([]EdgeCaseType, 0, len(parts))
	valid := make(map[EdgeCaseType]bool)
	for _, t := range AllEdgeCaseTypes() {
		valid[t] = true
	}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		t := EdgeCaseType(p)
		if !valid[t] {
			return nil, fmt.Errorf("unknown edge case type %q, valid types: %v", p, AllEdgeCaseTypes())
		}
		result = append(result, t)
	}
	return result, nil
}

// Validate checks if config is valid
func (c *Config) Validate() error {
	if c.Percentage < 0 || c.Percentage > 100 {
		return fmt.Errorf("edge-cases percentage must be 0-100, got %d", c.Percentage)
	}
	if c.Percentage > 0 && len(c.Types) == 0 {
		return fmt.Errorf("edge-cases enabled but no types specified")
	}
	return nil
}

// IsEnabled returns true if edge cases are enabled
func (c *Config) IsEnabled() bool {
	return c.Percentage > 0 && len(c.Types) > 0
}

// HasType checks if a specific edge case type is enabled
func (c *Config) HasType(t EdgeCaseType) bool {
	for _, ct := range c.Types {
		if ct == t {
			return true
		}
	}
	return false
}
