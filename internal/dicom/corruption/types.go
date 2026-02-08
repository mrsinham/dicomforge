package corruption

import (
	"fmt"
	"strings"
)

// CorruptionType represents a category of vendor-specific corruption
type CorruptionType string

const (
	SiemensCSA       CorruptionType = "siemens-csa"
	GEPrivate        CorruptionType = "ge-private"
	PhilipsPrivate   CorruptionType = "philips-private"
	MalformedLengths CorruptionType = "malformed-lengths"
)

// AllCorruptionTypes returns all valid corruption types
func AllCorruptionTypes() []CorruptionType {
	return []CorruptionType{SiemensCSA, GEPrivate, PhilipsPrivate, MalformedLengths}
}

// Config holds corruption generation settings
type Config struct {
	Types []CorruptionType
}

// ParseTypes parses comma-separated corruption types.
// The special value "all" enables all corruption types.
func ParseTypes(input string) ([]CorruptionType, error) {
	if input == "" {
		return nil, nil
	}

	parts := strings.Split(input, ",")
	valid := make(map[CorruptionType]bool)
	for _, t := range AllCorruptionTypes() {
		valid[t] = true
	}

	result := make([]CorruptionType, 0, len(parts))
	seen := make(map[CorruptionType]bool)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "all" {
			return AllCorruptionTypes(), nil
		}
		t := CorruptionType(p)
		if !valid[t] {
			return nil, fmt.Errorf("unknown corruption type %q, valid types: %v (or 'all')", p, AllCorruptionTypes())
		}
		if !seen[t] {
			result = append(result, t)
			seen[t] = true
		}
	}
	return result, nil
}

// Validate checks if config is valid
func (c *Config) Validate() error {
	if len(c.Types) == 0 {
		return fmt.Errorf("corruption enabled but no types specified")
	}
	valid := make(map[CorruptionType]bool)
	for _, t := range AllCorruptionTypes() {
		valid[t] = true
	}
	for _, t := range c.Types {
		if !valid[t] {
			return fmt.Errorf("unknown corruption type %q", t)
		}
	}
	return nil
}

// IsEnabled returns true if corruption is enabled
func (c *Config) IsEnabled() bool {
	return len(c.Types) > 0
}

// HasType checks if a specific corruption type is enabled
func (c *Config) HasType(t CorruptionType) bool {
	for _, ct := range c.Types {
		if ct == t {
			return true
		}
	}
	return false
}
