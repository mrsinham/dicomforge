package util

import (
	"strings"
	"testing"

	"github.com/suyashkumar/dicom/pkg/tag"
)

func TestGetTagByName_Valid(t *testing.T) {
	tests := []struct {
		name          string
		expectedTag   tag.Tag
		expectedScope TagScope
	}{
		// Patient level tags
		{"PatientName", tag.PatientName, ScopePatient},
		{"PatientID", tag.PatientID, ScopePatient},
		{"PatientBirthDate", tag.PatientBirthDate, ScopePatient},
		{"PatientSex", tag.PatientSex, ScopePatient},

		// Study level tags
		{"StudyDescription", tag.StudyDescription, ScopeStudy},
		{"InstitutionName", tag.InstitutionName, ScopeStudy},
		{"InstitutionalDepartmentName", tag.InstitutionalDepartmentName, ScopeStudy},
		{"ReferringPhysicianName", tag.ReferringPhysicianName, ScopeStudy},
		{"PerformingPhysicianName", tag.PerformingPhysicianName, ScopeStudy},
		{"OperatorsName", tag.OperatorsName, ScopeStudy},
		{"AccessionNumber", tag.AccessionNumber, ScopeStudy},
		{"StationName", tag.StationName, ScopeStudy},
		{"RequestedProcedurePriority", tag.RequestedProcedurePriority, ScopeStudy},
		{"RequestedProcedureDescription", tag.RequestedProcedureDescription, ScopeStudy},

		// Series level tags
		{"SeriesDescription", tag.SeriesDescription, ScopeSeries},
		{"ProtocolName", tag.ProtocolName, ScopeSeries},
		{"BodyPartExamined", tag.BodyPartExamined, ScopeSeries},
		{"SequenceName", tag.SequenceName, ScopeSeries},
		{"Manufacturer", tag.Manufacturer, ScopeSeries},
		{"ManufacturerModelName", tag.ManufacturerModelName, ScopeSeries},

		// Image level tags
		{"WindowCenter", tag.WindowCenter, ScopeImage},
		{"WindowWidth", tag.WindowWidth, ScopeImage},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			info, err := GetTagByName(tc.name)
			if err != nil {
				t.Fatalf("GetTagByName(%q) returned error: %v", tc.name, err)
			}
			if info.Tag != tc.expectedTag {
				t.Errorf("GetTagByName(%q).Tag = %v, want %v", tc.name, info.Tag, tc.expectedTag)
			}
			if info.Scope != tc.expectedScope {
				t.Errorf("GetTagByName(%q).Scope = %v, want %v", tc.name, info.Scope, tc.expectedScope)
			}
			if info.Name != tc.name {
				t.Errorf("GetTagByName(%q).Name = %q, want %q", tc.name, info.Name, tc.name)
			}
		})
	}
}

func TestGetTagByName_Invalid(t *testing.T) {
	invalidNames := []string{
		"InvalidTagName",
		"NotATag",
		"",
		"   ",
		"PatientNameXYZ",
	}

	for _, name := range invalidNames {
		t.Run(name, func(t *testing.T) {
			_, err := GetTagByName(name)
			if err == nil {
				t.Errorf("GetTagByName(%q) should return error for invalid tag", name)
			}
		})
	}
}

func TestGetTagByName_Suggestion(t *testing.T) {
	tests := []struct {
		typo       string
		suggestion string
	}{
		{"PatientNam", "PatientName"},
		{"PatinetName", "PatientName"},
		{"PatientNme", "PatientName"},
		{"StudyDescripton", "StudyDescription"},
		{"SeriesDescritpion", "SeriesDescription"},
		{"Manufacurer", "Manufacturer"},
		{"WindowCentre", "WindowCenter"},
	}

	for _, tc := range tests {
		t.Run(tc.typo, func(t *testing.T) {
			_, err := GetTagByName(tc.typo)
			if err == nil {
				t.Fatalf("GetTagByName(%q) should return error", tc.typo)
			}
			// Error message should contain the suggestion
			if !strings.Contains(err.Error(), tc.suggestion) {
				t.Errorf("Error for %q should suggest %q, got: %v", tc.typo, tc.suggestion, err)
			}
		})
	}
}

func TestGetTagByName_CaseInsensitive(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"patientname", "PatientName"},
		{"PATIENTNAME", "PatientName"},
		{"PatientNAME", "PatientName"},
		{"pAtIeNtNaMe", "PatientName"},
		{"studydescription", "StudyDescription"},
		{"STUDYDESCRIPTION", "StudyDescription"},
		{"windowcenter", "WindowCenter"},
		{"WINDOWWIDTH", "WindowWidth"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			info, err := GetTagByName(tc.input)
			if err != nil {
				t.Fatalf("GetTagByName(%q) returned error: %v", tc.input, err)
			}
			if info.Name != tc.expected {
				t.Errorf("GetTagByName(%q).Name = %q, want %q", tc.input, info.Name, tc.expected)
			}
		})
	}
}

func TestTagScope_String(t *testing.T) {
	tests := []struct {
		scope    TagScope
		expected string
	}{
		{ScopePatient, "Patient"},
		{ScopeStudy, "Study"},
		{ScopeSeries, "Series"},
		{ScopeImage, "Image"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			if tc.scope.String() != tc.expected {
				t.Errorf("TagScope.String() = %q, want %q", tc.scope.String(), tc.expected)
			}
		})
	}
}

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		a, b     string
		expected int
	}{
		{"", "", 0},
		{"abc", "", 3},
		{"", "abc", 3},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"kitten", "sitting", 3},
		{"saturday", "sunday", 3},
		{"PatientName", "PatinetName", 2}, // transposition counts as 2 in standard Levenshtein
	}

	for _, tc := range tests {
		t.Run(tc.a+"_"+tc.b, func(t *testing.T) {
			result := levenshteinDistance(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("levenshteinDistance(%q, %q) = %d, want %d", tc.a, tc.b, result, tc.expected)
			}
		})
	}
}
