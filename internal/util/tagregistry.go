// Package util provides utility functions for DICOM file generation.
package util

import (
	"fmt"
	"strings"

	"github.com/suyashkumar/dicom/pkg/tag"
)

// TagScope represents the DICOM hierarchy level at which a tag should be consistent.
type TagScope int

const (
	// ScopePatient indicates tags that should be consistent across all images for a patient.
	ScopePatient TagScope = iota
	// ScopeStudy indicates tags that should be consistent within a study.
	ScopeStudy
	// ScopeSeries indicates tags that should be consistent within a series.
	ScopeSeries
	// ScopeImage indicates tags that can vary per image.
	ScopeImage
)

// String returns the string representation of a TagScope.
func (s TagScope) String() string {
	switch s {
	case ScopePatient:
		return "Patient"
	case ScopeStudy:
		return "Study"
	case ScopeSeries:
		return "Series"
	case ScopeImage:
		return "Image"
	default:
		return "Unknown"
	}
}

// TagInfo contains information about a DICOM tag, including its scope.
type TagInfo struct {
	Name  string
	Tag   tag.Tag
	Scope TagScope
}

// tagRegistry maps lowercase tag names to their TagInfo.
var tagRegistry = map[string]TagInfo{
	// Patient level tags
	"patientname":      {Name: "PatientName", Tag: tag.PatientName, Scope: ScopePatient},
	"patientid":        {Name: "PatientID", Tag: tag.PatientID, Scope: ScopePatient},
	"patientbirthdate": {Name: "PatientBirthDate", Tag: tag.PatientBirthDate, Scope: ScopePatient},
	"patientsex":       {Name: "PatientSex", Tag: tag.PatientSex, Scope: ScopePatient},

	// Study level tags
	"studydescription":             {Name: "StudyDescription", Tag: tag.StudyDescription, Scope: ScopeStudy},
	"institutionname":              {Name: "InstitutionName", Tag: tag.InstitutionName, Scope: ScopeStudy},
	"institutionaldepartmentname":  {Name: "InstitutionalDepartmentName", Tag: tag.InstitutionalDepartmentName, Scope: ScopeStudy},
	"referringphysicianname":       {Name: "ReferringPhysicianName", Tag: tag.ReferringPhysicianName, Scope: ScopeStudy},
	"performingphysicianname":      {Name: "PerformingPhysicianName", Tag: tag.PerformingPhysicianName, Scope: ScopeStudy},
	"operatorsname":                {Name: "OperatorsName", Tag: tag.OperatorsName, Scope: ScopeStudy},
	"accessionnumber":              {Name: "AccessionNumber", Tag: tag.AccessionNumber, Scope: ScopeStudy},
	"stationname":                  {Name: "StationName", Tag: tag.StationName, Scope: ScopeStudy},
	"requestedprocedurepriority":   {Name: "RequestedProcedurePriority", Tag: tag.RequestedProcedurePriority, Scope: ScopeStudy},
	"requestedproceduredescription": {Name: "RequestedProcedureDescription", Tag: tag.RequestedProcedureDescription, Scope: ScopeStudy},

	// Series level tags
	"seriesdescription":     {Name: "SeriesDescription", Tag: tag.SeriesDescription, Scope: ScopeSeries},
	"protocolname":          {Name: "ProtocolName", Tag: tag.ProtocolName, Scope: ScopeSeries},
	"bodypartexamined":      {Name: "BodyPartExamined", Tag: tag.BodyPartExamined, Scope: ScopeSeries},
	"sequencename":          {Name: "SequenceName", Tag: tag.SequenceName, Scope: ScopeSeries},
	"manufacturer":          {Name: "Manufacturer", Tag: tag.Manufacturer, Scope: ScopeSeries},
	"manufacturermodelname": {Name: "ManufacturerModelName", Tag: tag.ManufacturerModelName, Scope: ScopeSeries},

	// Image level tags
	"windowcenter": {Name: "WindowCenter", Tag: tag.WindowCenter, Scope: ScopeImage},
	"windowwidth":  {Name: "WindowWidth", Tag: tag.WindowWidth, Scope: ScopeImage},
}

// GetTagByName returns TagInfo for a given tag name.
// The lookup is case-insensitive. If the tag is not found, an error is returned
// with a suggestion for the closest matching tag name (using Levenshtein distance).
func GetTagByName(name string) (TagInfo, error) {
	// Normalize the input name to lowercase
	normalizedName := strings.ToLower(strings.TrimSpace(name))

	// Direct lookup
	if info, ok := tagRegistry[normalizedName]; ok {
		return info, nil
	}

	// Tag not found, try to find a suggestion
	suggestion := findClosestTagName(normalizedName)
	if suggestion != "" {
		return TagInfo{}, fmt.Errorf("unknown tag %q, did you mean %q?", name, suggestion)
	}

	return TagInfo{}, fmt.Errorf("unknown tag %q", name)
}

// findClosestTagName finds the closest matching tag name using Levenshtein distance.
// Returns empty string if no close match is found (distance > 5).
func findClosestTagName(input string) string {
	const maxDistance = 5
	bestDistance := maxDistance + 1
	var bestMatch string

	for key, info := range tagRegistry {
		distance := levenshteinDistance(input, key)
		if distance < bestDistance {
			bestDistance = distance
			bestMatch = info.Name
		}
	}

	if bestDistance <= maxDistance {
		return bestMatch
	}
	return ""
}

// levenshteinDistance calculates the Levenshtein distance between two strings.
// This is the minimum number of single-character edits (insertions, deletions,
// or substitutions) required to change one string into the other.
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	// Create a matrix to store distances
	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
	}

	// Initialize the first row and column
	for i := 0; i <= len(a); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}

	// Fill in the rest of the matrix
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}
