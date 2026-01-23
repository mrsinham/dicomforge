# Generic Tag Option Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `--tag "TagName=Value"` CLI option to set arbitrary DICOM tags.

**Architecture:** Parse tag flags into a map, validate against known DICOM tags with fuzzy matching for typos, integrate with generator using tag scope rules (patient/study/series/image level).

**Tech Stack:** Go 1.24+, github.com/suyashkumar/dicom/pkg/tag, Levenshtein distance for fuzzy matching

---

## Task 1: Create Tag Registry with Scope Information

**Files:**
- Create: `internal/util/tagregistry.go`
- Test: `internal/util/tagregistry_test.go`

**Step 1: Write the failing test**

```go
// internal/util/tagregistry_test.go
package util

import "testing"

func TestGetTagByName_Valid(t *testing.T) {
	info, err := GetTagByName("InstitutionName")
	if err != nil {
		t.Fatalf("GetTagByName failed: %v", err)
	}
	if info.Name != "InstitutionName" {
		t.Errorf("Name = %s, want InstitutionName", info.Name)
	}
	if info.Scope != ScopeStudy {
		t.Errorf("Scope = %v, want ScopeStudy", info.Scope)
	}
}

func TestGetTagByName_Invalid(t *testing.T) {
	_, err := GetTagByName("NotARealTag")
	if err == nil {
		t.Error("Expected error for invalid tag")
	}
}

func TestGetTagByName_Suggestion(t *testing.T) {
	_, err := GetTagByName("InstitutioName") // typo
	if err == nil {
		t.Error("Expected error for typo")
	}
	// Error should contain suggestion
	errStr := err.Error()
	if !strings.Contains(errStr, "InstitutionName") {
		t.Errorf("Error should suggest InstitutionName, got: %s", errStr)
	}
}

func TestGetTagByName_CaseInsensitive(t *testing.T) {
	info, err := GetTagByName("institutionname")
	if err != nil {
		t.Fatalf("GetTagByName should be case-insensitive: %v", err)
	}
	if info.Name != "InstitutionName" {
		t.Errorf("Should return canonical name, got: %s", info.Name)
	}
}
```

**Step 2: Implement tag registry**

```go
// internal/util/tagregistry.go
package util

import (
	"fmt"
	"strings"

	"github.com/suyashkumar/dicom/pkg/tag"
)

// TagScope defines at which level a tag should be consistent
type TagScope int

const (
	ScopePatient TagScope = iota // Same value for all studies of a patient
	ScopeStudy                   // Same value within a study
	ScopeSeries                  // Same value within a series
	ScopeImage                   // Can vary per image
)

// TagInfo contains metadata about a DICOM tag
type TagInfo struct {
	Name  string
	Tag   tag.Tag
	Scope TagScope
}

// Registry of known tags with their scope
var tagRegistry = map[string]TagInfo{
	// Patient level
	"patientname":      {Name: "PatientName", Tag: tag.PatientName, Scope: ScopePatient},
	"patientid":        {Name: "PatientID", Tag: tag.PatientID, Scope: ScopePatient},
	"patientbirthdate": {Name: "PatientBirthDate", Tag: tag.PatientBirthDate, Scope: ScopePatient},
	"patientsex":       {Name: "PatientSex", Tag: tag.PatientSex, Scope: ScopePatient},

	// Study level
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

	// Series level
	"seriesdescription": {Name: "SeriesDescription", Tag: tag.SeriesDescription, Scope: ScopeSeries},
	"protocolname":      {Name: "ProtocolName", Tag: tag.ProtocolName, Scope: ScopeSeries},
	"bodypartexamined":  {Name: "BodyPartExamined", Tag: tag.BodyPartExamined, Scope: ScopeSeries},
	"sequencename":      {Name: "SequenceName", Tag: tag.SequenceName, Scope: ScopeSeries},
	"manufacturer":      {Name: "Manufacturer", Tag: tag.Manufacturer, Scope: ScopeSeries},
	"manufacturermodelname": {Name: "ManufacturerModelName", Tag: tag.ManufacturerModelName, Scope: ScopeSeries},

	// Image level
	"windowcenter": {Name: "WindowCenter", Tag: tag.WindowCenter, Scope: ScopeImage},
	"windowwidth":  {Name: "WindowWidth", Tag: tag.WindowWidth, Scope: ScopeImage},
}

// GetTagByName returns tag info by name (case-insensitive)
func GetTagByName(name string) (TagInfo, error) {
	lower := strings.ToLower(name)
	if info, ok := tagRegistry[lower]; ok {
		return info, nil
	}

	// Find closest match for suggestion
	closest := findClosestTag(lower)
	if closest != "" {
		return TagInfo{}, fmt.Errorf("unknown tag %q, did you mean %q?", name, closest)
	}
	return TagInfo{}, fmt.Errorf("unknown tag %q", name)
}

// findClosestTag finds the closest matching tag name using Levenshtein distance
func findClosestTag(input string) string {
	minDist := 999
	closest := ""
	for key, info := range tagRegistry {
		dist := levenshtein(input, key)
		if dist < minDist && dist <= 3 { // Max 3 edits
			minDist = dist
			closest = info.Name
		}
	}
	return closest
}

// levenshtein calculates the edit distance between two strings
func levenshtein(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
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

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
```

**Step 3: Commit**

```bash
git add internal/util/tagregistry.go internal/util/tagregistry_test.go
git commit -m "feat: add tag registry with scope and fuzzy matching"
```

---

## Task 2: Add Tag Parsing Function

**Files:**
- Create: `internal/util/tagparser.go`
- Test: `internal/util/tagparser_test.go`

**Step 1: Write the failing test**

```go
// internal/util/tagparser_test.go
package util

import "testing"

func TestParseTagFlag_Valid(t *testing.T) {
	tags, err := ParseTagFlags([]string{
		"InstitutionName=CHU Bordeaux",
		"BodyPartExamined=HEAD",
	})
	if err != nil {
		t.Fatalf("ParseTagFlags failed: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("Expected 2 tags, got %d", len(tags))
	}
	if tags["InstitutionName"] != "CHU Bordeaux" {
		t.Errorf("InstitutionName = %s, want CHU Bordeaux", tags["InstitutionName"])
	}
}

func TestParseTagFlag_InvalidFormat(t *testing.T) {
	_, err := ParseTagFlags([]string{"NoEqualsSign"})
	if err == nil {
		t.Error("Expected error for missing '='")
	}
}

func TestParseTagFlag_UnknownTag(t *testing.T) {
	_, err := ParseTagFlags([]string{"FakeTag=Value"})
	if err == nil {
		t.Error("Expected error for unknown tag")
	}
}

func TestParseTagFlag_EmptyValue(t *testing.T) {
	tags, err := ParseTagFlags([]string{"InstitutionName="})
	if err != nil {
		t.Fatalf("Empty value should be allowed: %v", err)
	}
	if tags["InstitutionName"] != "" {
		t.Error("Empty value should be preserved")
	}
}
```

**Step 2: Implement parser**

```go
// internal/util/tagparser.go
package util

import (
	"fmt"
	"strings"
)

// ParsedTags maps canonical tag names to their values
type ParsedTags map[string]string

// ParseTagFlags parses --tag flags into a validated map
func ParseTagFlags(flags []string) (ParsedTags, error) {
	result := make(ParsedTags)

	for _, flag := range flags {
		idx := strings.Index(flag, "=")
		if idx == -1 {
			return nil, fmt.Errorf("invalid tag format %q: expected 'TagName=Value'", flag)
		}

		name := flag[:idx]
		value := flag[idx+1:]

		info, err := GetTagByName(name)
		if err != nil {
			return nil, err
		}

		result[info.Name] = value
	}

	return result, nil
}

// GetWithScope returns tags filtered by scope
func (p ParsedTags) GetWithScope(scope TagScope) ParsedTags {
	result := make(ParsedTags)
	for name, value := range p {
		info, _ := GetTagByName(name)
		if info.Scope == scope {
			result[name] = value
		}
	}
	return result
}

// Has checks if a tag is defined
func (p ParsedTags) Has(name string) bool {
	_, ok := p[name]
	return ok
}

// Get returns the value or empty string
func (p ParsedTags) Get(name string) string {
	return p[name]
}
```

**Step 3: Commit**

```bash
git add internal/util/tagparser.go internal/util/tagparser_test.go
git commit -m "feat: add tag flag parser with validation"
```

---

## Task 3: Add --tag CLI Flag

**Files:**
- Modify: `cmd/dicomforge/main.go`

**Changes:**

1. Add new flag that can be repeated:
```go
var tagFlags []string
flag.Func("tag", "Set DICOM tag: 'TagName=Value' (repeatable)", func(s string) error {
	tagFlags = append(tagFlags, s)
	return nil
})
```

2. Parse tags after flag.Parse():
```go
parsedTags, err := util.ParseTagFlags(tagFlags)
if err != nil {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
```

3. Add to GeneratorOptions:
```go
opts := dicom.GeneratorOptions{
	// ... existing fields ...
	CustomTags: parsedTags,
}
```

4. Update help text

**Step: Commit**

```bash
git add cmd/dicomforge/main.go
git commit -m "feat: add --tag CLI flag for custom DICOM tags"
```

---

## Task 4: Update Generator to Use Custom Tags

**Files:**
- Modify: `internal/dicom/generator.go`

**Changes:**

1. Add CustomTags to GeneratorOptions:
```go
type GeneratorOptions struct {
	// ... existing fields ...
	CustomTags util.ParsedTags // User-defined tag overrides
}
```

2. Create helper to get tag value (custom or generated):
```go
func (opts *GeneratorOptions) getTagValue(name, generated string) string {
	if opts.CustomTags.Has(name) {
		return opts.CustomTags.Get(name)
	}
	return generated
}
```

3. Update metadata generation to use helper:
```go
// Instead of:
mustNewElement(tag.InstitutionName, []string{studyInstitution.Name}),

// Use:
mustNewElement(tag.InstitutionName, []string{opts.getTagValue("InstitutionName", studyInstitution.Name)}),
```

4. Apply to all overridable tags at appropriate scope levels

**Step: Commit**

```bash
git add internal/dicom/generator.go
git commit -m "feat: integrate custom tags with generator"
```

---

## Task 5: Add Integration Test

**Files:**
- Modify: `tests/integration_test.go`

**Test:**
```go
func TestCustomTags(t *testing.T) {
	tmpDir := t.TempDir()

	customTags, _ := util.ParseTagFlags([]string{
		"InstitutionName=Custom Hospital",
		"ReferringPhysicianName=Dr Custom^Name",
	})

	opts := dicom.GeneratorOptions{
		NumImages:   2,
		TotalSize:   "1MB",
		OutputDir:   tmpDir,
		Seed:        42,
		NumStudies:  1,
		NumPatients: 1,
		CustomTags:  customTags,
	}

	files, err := dicom.GenerateDICOMSeries(opts)
	// ... verify custom tags are in output ...
}
```

**Step: Commit**

```bash
git add tests/integration_test.go
git commit -m "test: add integration test for custom tags"
```

---

## Summary

5 tasks:
1. Tag registry with scope and fuzzy matching
2. Tag flag parser with validation
3. CLI --tag flag
4. Generator integration
5. Integration test

**New CLI usage:**
```bash
dicomforge --num-images 10 --total-size 100MB \
  --tag "InstitutionName=CHU Bordeaux" \
  --tag "ReferringPhysicianName=Dr Martin^Jean" \
  --tag "StudyDescription=IRM Cerebrale"
```
