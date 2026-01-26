package wizard

import (
	"testing"

	"github.com/mrsinham/dicomforge/cmd/dicomforge/wizard/types"
	"github.com/mrsinham/dicomforge/internal/dicom/modalities"
	"github.com/mrsinham/dicomforge/internal/util"
)

func TestToGeneratorOptions_BasicConversion(t *testing.T) {
	state := &WizardState{
		Global: types.GlobalConfig{
			Modality:          "MR",
			TotalImages:       100,
			TotalSize:         "1GB",
			OutputDir:         "/output/dir",
			Seed:              12345,
			NumPatients:       2,
			StudiesPerPatient: 2,
			SeriesPerStudy:    3,
		},
		Patients: []types.PatientConfig{
			{
				Name:      "Patient One",
				ID:        "P001",
				BirthDate: "1990-01-01",
				Sex:       "M",
				Studies: []types.StudyConfig{
					{
						Description:     "Brain MRI",
						Date:            "2024-01-01",
						AccessionNumber: "ACC001",
						Institution:     "Test Hospital",
						Department:      "Radiology",
						BodyPart:        "HEAD",
						Priority:        "ROUTINE",
						Series: []types.SeriesConfig{
							{Description: "T1", ImageCount: 20},
							{Description: "T2", ImageCount: 20},
						},
					},
					{
						Description: "Spine MRI",
						BodyPart:    "SPINE",
						Series: []types.SeriesConfig{
							{Description: "Sagittal", ImageCount: 30},
						},
					},
				},
			},
			{
				Name: "Patient Two",
				ID:   "P002",
				Studies: []types.StudyConfig{
					{
						Description: "Chest MRI",
						BodyPart:    "CHEST",
						Series: []types.SeriesConfig{
							{Description: "Axial", ImageCount: 30},
						},
					},
				},
			},
		},
	}

	wizard := &Wizard{state: state}
	opts, err := wizard.toGeneratorOptions()
	if err != nil {
		t.Fatalf("toGeneratorOptions failed: %v", err)
	}

	// Verify basic fields
	if opts.NumImages != 100 {
		t.Errorf("Expected NumImages 100, got %d", opts.NumImages)
	}
	if opts.TotalSize != "1GB" {
		t.Errorf("Expected TotalSize 1GB, got %s", opts.TotalSize)
	}
	if opts.OutputDir != "/output/dir" {
		t.Errorf("Expected OutputDir /output/dir, got %s", opts.OutputDir)
	}
	if opts.Seed != 12345 {
		t.Errorf("Expected Seed 12345, got %d", opts.Seed)
	}
	if opts.NumPatients != 2 {
		t.Errorf("Expected NumPatients 2, got %d", opts.NumPatients)
	}

	// Verify modality
	if opts.Modality != modalities.MR {
		t.Errorf("Expected Modality MR, got %s", opts.Modality)
	}

	// Verify study count (total studies from all patients)
	expectedStudies := 3 // 2 for patient 1 + 1 for patient 2
	if opts.NumStudies != expectedStudies {
		t.Errorf("Expected NumStudies %d, got %d", expectedStudies, opts.NumStudies)
	}

	// Verify series per study
	if opts.SeriesPerStudy.Min != 3 || opts.SeriesPerStudy.Max != 3 {
		t.Errorf("Expected SeriesPerStudy {3,3}, got {%d,%d}", opts.SeriesPerStudy.Min, opts.SeriesPerStudy.Max)
	}

	// Verify study descriptions are collected
	if len(opts.StudyDescriptions) != 3 {
		t.Errorf("Expected 3 study descriptions, got %d", len(opts.StudyDescriptions))
	}
	if opts.StudyDescriptions[0] != "Brain MRI" {
		t.Errorf("Expected first study description 'Brain MRI', got %s", opts.StudyDescriptions[0])
	}

	// Verify body part, institution, department from first study
	if opts.BodyPart != "HEAD" {
		t.Errorf("Expected BodyPart HEAD, got %s", opts.BodyPart)
	}
	if opts.Institution != "Test Hospital" {
		t.Errorf("Expected Institution 'Test Hospital', got %s", opts.Institution)
	}
	if opts.Department != "Radiology" {
		t.Errorf("Expected Department 'Radiology', got %s", opts.Department)
	}
}

func TestToGeneratorOptions_CTModality(t *testing.T) {
	state := &WizardState{
		Global: types.GlobalConfig{
			Modality:          "CT",
			TotalImages:       50,
			TotalSize:         "500MB",
			OutputDir:         "/ct/output",
			NumPatients:       1,
			StudiesPerPatient: 1,
			SeriesPerStudy:    1,
		},
		Patients: []types.PatientConfig{
			{
				Name: "CT Patient",
				Studies: []types.StudyConfig{
					{
						Description: "CT Chest",
						BodyPart:    "CHEST",
						Series: []types.SeriesConfig{
							{Description: "Helical", ImageCount: 50},
						},
					},
				},
			},
		},
	}

	wizard := &Wizard{state: state}
	opts, err := wizard.toGeneratorOptions()
	if err != nil {
		t.Fatalf("toGeneratorOptions failed: %v", err)
	}

	if opts.Modality != modalities.CT {
		t.Errorf("Expected Modality CT, got %s", opts.Modality)
	}
}

func TestToGeneratorOptions_InvalidModality(t *testing.T) {
	state := &WizardState{
		Global: types.GlobalConfig{
			Modality:          "INVALID_MODALITY",
			TotalImages:       10,
			TotalSize:         "100MB",
			OutputDir:         "/output",
			NumPatients:       1,
			StudiesPerPatient: 1,
			SeriesPerStudy:    1,
		},
		Patients: []types.PatientConfig{},
	}

	wizard := &Wizard{state: state}
	opts, err := wizard.toGeneratorOptions()
	if err != nil {
		t.Fatalf("toGeneratorOptions should not fail for invalid modality: %v", err)
	}

	// Should default to MR for invalid modality
	if opts.Modality != modalities.MR {
		t.Errorf("Expected default Modality MR for invalid input, got %s", opts.Modality)
	}
}

func TestToGeneratorOptions_CustomTags(t *testing.T) {
	state := &WizardState{
		Global: types.GlobalConfig{
			Modality:          "MR",
			TotalImages:       20,
			TotalSize:         "200MB",
			OutputDir:         "/output",
			NumPatients:       1,
			StudiesPerPatient: 1,
			SeriesPerStudy:    1,
		},
		Patients: []types.PatientConfig{
			{
				Name: "Test Patient",
				Studies: []types.StudyConfig{
					{
						Description: "Test Study",
						CustomTags: map[string]string{
							"StudyComments":      "Study level tag",
							"InstitutionAddress": "123 Medical Way",
						},
						Series: []types.SeriesConfig{
							{
								Description: "Test Series",
								ImageCount:  20,
								CustomTags: map[string]string{
									"SeriesComments": "Series level tag",
									"ProtocolName":   "Custom Protocol",
								},
							},
						},
					},
				},
			},
		},
	}

	wizard := &Wizard{state: state}
	opts, err := wizard.toGeneratorOptions()
	if err != nil {
		t.Fatalf("toGeneratorOptions failed: %v", err)
	}

	// Verify custom tags are aggregated
	if opts.CustomTags == nil {
		t.Fatal("Expected CustomTags to be set")
	}

	// Check study-level tags
	if val, ok := opts.CustomTags.Get("StudyComments"); !ok || val != "Study level tag" {
		t.Errorf("Study-level custom tag not found or incorrect")
	}
	if val, ok := opts.CustomTags.Get("InstitutionAddress"); !ok || val != "123 Medical Way" {
		t.Errorf("InstitutionAddress custom tag not found or incorrect")
	}

	// Check series-level tags
	if val, ok := opts.CustomTags.Get("SeriesComments"); !ok || val != "Series level tag" {
		t.Errorf("Series-level custom tag not found or incorrect")
	}
	if val, ok := opts.CustomTags.Get("ProtocolName"); !ok || val != "Custom Protocol" {
		t.Errorf("ProtocolName custom tag not found or incorrect")
	}
}

func TestToGeneratorOptions_EmptyPatients(t *testing.T) {
	state := &WizardState{
		Global: types.GlobalConfig{
			Modality:          "MR",
			TotalImages:       30,
			TotalSize:         "300MB",
			OutputDir:         "/output",
			NumPatients:       3,
			StudiesPerPatient: 2,
			SeriesPerStudy:    1,
		},
		Patients: []types.PatientConfig{}, // No patients configured
	}

	wizard := &Wizard{state: state}
	opts, err := wizard.toGeneratorOptions()
	if err != nil {
		t.Fatalf("toGeneratorOptions failed: %v", err)
	}

	// When no patients are configured, it should use NumPatients * StudiesPerPatient
	expectedStudies := 3 * 2 // NumPatients * StudiesPerPatient
	if opts.NumStudies != expectedStudies {
		t.Errorf("Expected NumStudies %d for empty patients, got %d", expectedStudies, opts.NumStudies)
	}
}

func TestToGeneratorOptions_ZeroSeriesPerStudy(t *testing.T) {
	state := &WizardState{
		Global: types.GlobalConfig{
			Modality:          "MR",
			TotalImages:       10,
			TotalSize:         "100MB",
			OutputDir:         "/output",
			NumPatients:       1,
			StudiesPerPatient: 1,
			SeriesPerStudy:    0, // Zero should default to 1
		},
		Patients: []types.PatientConfig{
			{
				Name: "Test",
				Studies: []types.StudyConfig{
					{
						Description: "Test Study",
						Series:      []types.SeriesConfig{{Description: "S1", ImageCount: 10}},
					},
				},
			},
		},
	}

	wizard := &Wizard{state: state}
	opts, err := wizard.toGeneratorOptions()
	if err != nil {
		t.Fatalf("toGeneratorOptions failed: %v", err)
	}

	// SeriesPerStudy <= 0 should default to 1
	if opts.SeriesPerStudy.Min != 1 || opts.SeriesPerStudy.Max != 1 {
		t.Errorf("Expected SeriesPerStudy {1,1} for zero input, got {%d,%d}",
			opts.SeriesPerStudy.Min, opts.SeriesPerStudy.Max)
	}
}

func TestToGeneratorOptions_AllModalities(t *testing.T) {
	modalityTests := []struct {
		input    string
		expected modalities.Modality
	}{
		{"MR", modalities.MR},
		{"CT", modalities.CT},
		{"US", modalities.US},
		{"CR", modalities.CR},
		{"DX", modalities.DX},
		{"MG", modalities.MG},
	}

	for _, tc := range modalityTests {
		t.Run(tc.input, func(t *testing.T) {
			state := &WizardState{
				Global: types.GlobalConfig{
					Modality:          tc.input,
					TotalImages:       10,
					TotalSize:         "100MB",
					OutputDir:         "/output",
					NumPatients:       1,
					StudiesPerPatient: 1,
					SeriesPerStudy:    1,
				},
				Patients: []types.PatientConfig{
					{
						Name: "Test",
						Studies: []types.StudyConfig{
							{
								Description: "Test Study",
								Series:      []types.SeriesConfig{{Description: "S1", ImageCount: 10}},
							},
						},
					},
				},
			}

			wizard := &Wizard{state: state}
			opts, err := wizard.toGeneratorOptions()
			if err != nil {
				t.Fatalf("toGeneratorOptions failed for modality %s: %v", tc.input, err)
			}

			if opts.Modality != tc.expected {
				t.Errorf("Expected modality %s, got %s", tc.expected, opts.Modality)
			}
		})
	}
}

func TestToGeneratorOptions_MultipleStudiesCollectsDescriptions(t *testing.T) {
	state := &WizardState{
		Global: types.GlobalConfig{
			Modality:          "MR",
			TotalImages:       50,
			TotalSize:         "500MB",
			OutputDir:         "/output",
			NumPatients:       1,
			StudiesPerPatient: 3,
			SeriesPerStudy:    1,
		},
		Patients: []types.PatientConfig{
			{
				Name: "Multi-Study Patient",
				Studies: []types.StudyConfig{
					{Description: "Study Alpha", Series: []types.SeriesConfig{{Description: "S1", ImageCount: 15}}},
					{Description: "Study Beta", Series: []types.SeriesConfig{{Description: "S2", ImageCount: 15}}},
					{Description: "Study Gamma", Series: []types.SeriesConfig{{Description: "S3", ImageCount: 20}}},
				},
			},
		},
	}

	wizard := &Wizard{state: state}
	opts, err := wizard.toGeneratorOptions()
	if err != nil {
		t.Fatalf("toGeneratorOptions failed: %v", err)
	}

	if len(opts.StudyDescriptions) != 3 {
		t.Fatalf("Expected 3 study descriptions, got %d", len(opts.StudyDescriptions))
	}

	expectedDescriptions := []string{"Study Alpha", "Study Beta", "Study Gamma"}
	for i, expected := range expectedDescriptions {
		if opts.StudyDescriptions[i] != expected {
			t.Errorf("Study description %d: expected %q, got %q", i, expected, opts.StudyDescriptions[i])
		}
	}
}

func TestToGeneratorOptions_SeriesRangeFormat(t *testing.T) {
	state := &WizardState{
		Global: types.GlobalConfig{
			Modality:          "MR",
			TotalImages:       100,
			TotalSize:         "1GB",
			OutputDir:         "/output",
			NumPatients:       1,
			StudiesPerPatient: 1,
			SeriesPerStudy:    5,
		},
		Patients: []types.PatientConfig{
			{
				Name: "Test",
				Studies: []types.StudyConfig{
					{
						Description: "Test",
						Series:      []types.SeriesConfig{{Description: "S1", ImageCount: 100}},
					},
				},
			},
		},
	}

	wizard := &Wizard{state: state}
	opts, err := wizard.toGeneratorOptions()
	if err != nil {
		t.Fatalf("toGeneratorOptions failed: %v", err)
	}

	// SeriesPerStudy should be set as a fixed range (same min and max)
	expected := util.SeriesRange{Min: 5, Max: 5}
	if opts.SeriesPerStudy != expected {
		t.Errorf("Expected SeriesPerStudy %+v, got %+v", expected, opts.SeriesPerStudy)
	}
}

func TestNewWizard_DefaultState(t *testing.T) {
	wizard := NewWizard(nil)

	if wizard.state == nil {
		t.Fatal("Expected wizard.state to be initialized")
	}

	// Check default values
	if wizard.state.Global.Modality != "MR" {
		t.Errorf("Expected default modality MR, got %s", wizard.state.Global.Modality)
	}
	if wizard.state.Global.TotalImages != 50 {
		t.Errorf("Expected default total_images 50, got %d", wizard.state.Global.TotalImages)
	}
	if wizard.state.Global.TotalSize != "500MB" {
		t.Errorf("Expected default total_size 500MB, got %s", wizard.state.Global.TotalSize)
	}
	if wizard.state.Global.OutputDir != "dicom_series" {
		t.Errorf("Expected default output dicom_series, got %s", wizard.state.Global.OutputDir)
	}
	if wizard.state.Global.NumPatients != 1 {
		t.Errorf("Expected default num_patients 1, got %d", wizard.state.Global.NumPatients)
	}
	if wizard.state.Global.StudiesPerPatient != 1 {
		t.Errorf("Expected default studies_per_patient 1, got %d", wizard.state.Global.StudiesPerPatient)
	}
	if wizard.state.Global.SeriesPerStudy != 1 {
		t.Errorf("Expected default series_per_study 1, got %d", wizard.state.Global.SeriesPerStudy)
	}

	// Check initial phase
	if wizard.phase != PhaseGlobal {
		t.Errorf("Expected initial phase PhaseGlobal, got %v", wizard.phase)
	}
}

func TestNewWizard_WithExistingState(t *testing.T) {
	existingState := &WizardState{
		Global: types.GlobalConfig{
			Modality:          "CT",
			TotalImages:       200,
			TotalSize:         "2GB",
			OutputDir:         "/custom/path",
			NumPatients:       5,
			StudiesPerPatient: 3,
			SeriesPerStudy:    4,
		},
	}

	wizard := NewWizard(existingState)

	if wizard.state != existingState {
		t.Error("Expected wizard to use provided state")
	}
	if wizard.state.Global.Modality != "CT" {
		t.Errorf("Expected modality CT, got %s", wizard.state.Global.Modality)
	}
	if wizard.state.Global.TotalImages != 200 {
		t.Errorf("Expected total_images 200, got %d", wizard.state.Global.TotalImages)
	}
}
