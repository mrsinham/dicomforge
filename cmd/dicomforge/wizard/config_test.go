package wizard

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/mrsinham/dicomforge/cmd/dicomforge/wizard/types"
)

func TestLoadFromYAML_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.yaml")

	content := `
global:
  modality: MR
  total_images: 10
  total_size: 100MB
  output: ./test
  seed: 42
  num_patients: 2
  studies_per_patient: 3
  series_per_study: 2
patients:
  - name: "John Doe"
    id: "PAT001"
    birth_date: "1990-01-15"
    sex: "M"
    studies:
      - description: "Brain MRI"
        date: "2024-01-01"
        accession: "ACC001"
        institution: "Test Hospital"
        department: "Radiology"
        body_part: "HEAD"
        priority: "ROUTINE"
        referring_physician: "Dr. Smith"
        custom_tags:
          PatientComments: "Test comment"
        series:
          - description: "T1 Weighted"
            protocol: "T1W"
            orientation: "AXIAL"
            images: 50
            custom_tags:
              SeriesDescription: "Custom Series"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	state, err := LoadFromYAML(configPath)
	if err != nil {
		t.Fatalf("LoadFromYAML failed: %v", err)
	}

	// Verify global config
	if state.Global.Modality != "MR" {
		t.Errorf("Expected modality MR, got %s", state.Global.Modality)
	}
	if state.Global.TotalImages != 10 {
		t.Errorf("Expected total_images 10, got %d", state.Global.TotalImages)
	}
	if state.Global.TotalSize != "100MB" {
		t.Errorf("Expected total_size 100MB, got %s", state.Global.TotalSize)
	}
	if state.Global.OutputDir != "./test" {
		t.Errorf("Expected output ./test, got %s", state.Global.OutputDir)
	}
	if state.Global.Seed != 42 {
		t.Errorf("Expected seed 42, got %d", state.Global.Seed)
	}
	if state.Global.NumPatients != 2 {
		t.Errorf("Expected num_patients 2, got %d", state.Global.NumPatients)
	}
	if state.Global.StudiesPerPatient != 3 {
		t.Errorf("Expected studies_per_patient 3, got %d", state.Global.StudiesPerPatient)
	}
	if state.Global.SeriesPerStudy != 2 {
		t.Errorf("Expected series_per_study 2, got %d", state.Global.SeriesPerStudy)
	}

	// Verify patient config
	if len(state.Patients) != 1 {
		t.Fatalf("Expected 1 patient, got %d", len(state.Patients))
	}
	patient := state.Patients[0]
	if patient.Name != "John Doe" {
		t.Errorf("Expected patient name 'John Doe', got %s", patient.Name)
	}
	if patient.ID != "PAT001" {
		t.Errorf("Expected patient ID 'PAT001', got %s", patient.ID)
	}
	if patient.BirthDate != "1990-01-15" {
		t.Errorf("Expected birth_date '1990-01-15', got %s", patient.BirthDate)
	}
	if patient.Sex != "M" {
		t.Errorf("Expected sex 'M', got %s", patient.Sex)
	}

	// Verify study config
	if len(patient.Studies) != 1 {
		t.Fatalf("Expected 1 study, got %d", len(patient.Studies))
	}
	study := patient.Studies[0]
	if study.Description != "Brain MRI" {
		t.Errorf("Expected study description 'Brain MRI', got %s", study.Description)
	}
	if study.Date != "2024-01-01" {
		t.Errorf("Expected study date '2024-01-01', got %s", study.Date)
	}
	if study.AccessionNumber != "ACC001" {
		t.Errorf("Expected accession 'ACC001', got %s", study.AccessionNumber)
	}
	if study.Institution != "Test Hospital" {
		t.Errorf("Expected institution 'Test Hospital', got %s", study.Institution)
	}
	if study.CustomTags["PatientComments"] != "Test comment" {
		t.Errorf("Expected custom tag 'Test comment', got %s", study.CustomTags["PatientComments"])
	}

	// Verify series config
	if len(study.Series) != 1 {
		t.Fatalf("Expected 1 series, got %d", len(study.Series))
	}
	series := study.Series[0]
	if series.Description != "T1 Weighted" {
		t.Errorf("Expected series description 'T1 Weighted', got %s", series.Description)
	}
	if series.Protocol != "T1W" {
		t.Errorf("Expected protocol 'T1W', got %s", series.Protocol)
	}
	if series.Orientation != "AXIAL" {
		t.Errorf("Expected orientation 'AXIAL', got %s", series.Orientation)
	}
	if series.ImageCount != 50 {
		t.Errorf("Expected image count 50, got %d", series.ImageCount)
	}
	if series.CustomTags["SeriesDescription"] != "Custom Series" {
		t.Errorf("Expected custom tag 'Custom Series', got %s", series.CustomTags["SeriesDescription"])
	}
}

func TestLoadFromYAML_NonExistentFile(t *testing.T) {
	_, err := LoadFromYAML("/non/existent/path/config.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestLoadFromYAML_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	// Invalid YAML content
	content := `
global:
  modality: MR
  total_images: [invalid array in scalar field
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, err := LoadFromYAML(configPath)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestSaveToYAML_AndLoadBack(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "output.yaml")

	state := &WizardState{
		Global: types.GlobalConfig{
			Modality:          "CT",
			TotalImages:       100,
			TotalSize:         "1GB",
			OutputDir:         "/output/path",
			Seed:              12345,
			NumPatients:       3,
			StudiesPerPatient: 2,
			SeriesPerStudy:    4,
		},
		Patients: []types.PatientConfig{
			{
				Name:      "Jane Smith",
				ID:        "PAT002",
				BirthDate: "1985-06-20",
				Sex:       "F",
				Studies: []types.StudyConfig{
					{
						Description:        "Chest CT",
						Date:               "2024-02-15",
						AccessionNumber:    "ACC002",
						Institution:        "City Hospital",
						Department:         "Emergency",
						BodyPart:           "CHEST",
						Priority:           "STAT",
						ReferringPhysician: "Dr. Jones",
						CustomTags: map[string]string{
							"StudyComments": "Urgent scan",
						},
						Series: []types.SeriesConfig{
							{
								Description: "Helical",
								Protocol:    "HELICAL",
								Orientation: "AXIAL",
								ImageCount:  200,
								CustomTags: map[string]string{
									"ProtocolName": "CT Chest Protocol",
								},
							},
						},
					},
				},
			},
		},
	}

	// Save to YAML
	if err := SaveToYAML(state, configPath); err != nil {
		t.Fatalf("SaveToYAML failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load it back
	loaded, err := LoadFromYAML(configPath)
	if err != nil {
		t.Fatalf("LoadFromYAML failed: %v", err)
	}

	// Verify global config matches
	if loaded.Global.Modality != state.Global.Modality {
		t.Errorf("Modality mismatch: expected %s, got %s", state.Global.Modality, loaded.Global.Modality)
	}
	if loaded.Global.TotalImages != state.Global.TotalImages {
		t.Errorf("TotalImages mismatch: expected %d, got %d", state.Global.TotalImages, loaded.Global.TotalImages)
	}
	if loaded.Global.TotalSize != state.Global.TotalSize {
		t.Errorf("TotalSize mismatch: expected %s, got %s", state.Global.TotalSize, loaded.Global.TotalSize)
	}
	if loaded.Global.OutputDir != state.Global.OutputDir {
		t.Errorf("OutputDir mismatch: expected %s, got %s", state.Global.OutputDir, loaded.Global.OutputDir)
	}
	if loaded.Global.Seed != state.Global.Seed {
		t.Errorf("Seed mismatch: expected %d, got %d", state.Global.Seed, loaded.Global.Seed)
	}

	// Verify patient matches
	if len(loaded.Patients) != len(state.Patients) {
		t.Fatalf("Patient count mismatch: expected %d, got %d", len(state.Patients), len(loaded.Patients))
	}
	if loaded.Patients[0].Name != state.Patients[0].Name {
		t.Errorf("Patient name mismatch: expected %s, got %s", state.Patients[0].Name, loaded.Patients[0].Name)
	}

	// Verify study matches
	if len(loaded.Patients[0].Studies) != len(state.Patients[0].Studies) {
		t.Fatalf("Study count mismatch")
	}
	if loaded.Patients[0].Studies[0].Description != state.Patients[0].Studies[0].Description {
		t.Errorf("Study description mismatch")
	}
	if loaded.Patients[0].Studies[0].CustomTags["StudyComments"] != "Urgent scan" {
		t.Errorf("Study custom tag mismatch")
	}

	// Verify series matches
	if len(loaded.Patients[0].Studies[0].Series) != len(state.Patients[0].Studies[0].Series) {
		t.Fatalf("Series count mismatch")
	}
	if loaded.Patients[0].Studies[0].Series[0].ImageCount != state.Patients[0].Studies[0].Series[0].ImageCount {
		t.Errorf("Series image count mismatch")
	}
	if loaded.Patients[0].Studies[0].Series[0].CustomTags["ProtocolName"] != "CT Chest Protocol" {
		t.Errorf("Series custom tag mismatch")
	}
}

func TestConfigToWizardState(t *testing.T) {
	cfg := &Config{
		Global: GlobalConfigYAML{
			Modality:          "US",
			TotalImages:       50,
			TotalSize:         "500MB",
			OutputDir:         "/dicom/output",
			Seed:              999,
			NumPatients:       5,
			StudiesPerPatient: 1,
			SeriesPerStudy:    3,
		},
		Patients: []PatientConfigYAML{
			{
				Name:      "Test Patient",
				ID:        "TP001",
				BirthDate: "2000-01-01",
				Sex:       "O",
				Studies: []StudyConfigYAML{
					{
						Description:        "Ultrasound Abdomen",
						Date:               "2024-03-01",
						AccessionNumber:    "US001",
						Institution:        "Imaging Center",
						Department:         "Ultrasound",
						BodyPart:           "ABDOMEN",
						Priority:           "ROUTINE",
						ReferringPhysician: "Dr. Brown",
						CustomTags: map[string]string{
							"PerformingPhysiciansName": "Dr. Operator",
						},
						Series: []SeriesConfigYAML{
							{
								Description: "Liver",
								Protocol:    "LIVER_PROTOCOL",
								Orientation: "TRANSVERSE",
								ImageCount:  30,
								CustomTags: map[string]string{
									"BodyPartExamined": "LIVER",
								},
							},
						},
					},
				},
			},
		},
	}

	state := configToWizardState(cfg)

	// Verify global conversion
	if state.Global.Modality != cfg.Global.Modality {
		t.Errorf("Modality not converted correctly")
	}
	if state.Global.TotalImages != cfg.Global.TotalImages {
		t.Errorf("TotalImages not converted correctly")
	}
	if state.Global.TotalSize != cfg.Global.TotalSize {
		t.Errorf("TotalSize not converted correctly")
	}
	if state.Global.OutputDir != cfg.Global.OutputDir {
		t.Errorf("OutputDir not converted correctly")
	}
	if state.Global.Seed != cfg.Global.Seed {
		t.Errorf("Seed not converted correctly")
	}
	if state.Global.NumPatients != cfg.Global.NumPatients {
		t.Errorf("NumPatients not converted correctly")
	}
	if state.Global.StudiesPerPatient != cfg.Global.StudiesPerPatient {
		t.Errorf("StudiesPerPatient not converted correctly")
	}
	if state.Global.SeriesPerStudy != cfg.Global.SeriesPerStudy {
		t.Errorf("SeriesPerStudy not converted correctly")
	}

	// Verify patient conversion
	if len(state.Patients) != 1 {
		t.Fatalf("Expected 1 patient, got %d", len(state.Patients))
	}
	if state.Patients[0].Name != "Test Patient" {
		t.Errorf("Patient name not converted correctly")
	}
	if state.Patients[0].ID != "TP001" {
		t.Errorf("Patient ID not converted correctly")
	}

	// Verify study conversion
	if len(state.Patients[0].Studies) != 1 {
		t.Fatalf("Expected 1 study")
	}
	study := state.Patients[0].Studies[0]
	if study.Description != "Ultrasound Abdomen" {
		t.Errorf("Study description not converted correctly")
	}
	if study.CustomTags["PerformingPhysiciansName"] != "Dr. Operator" {
		t.Errorf("Study custom tags not converted correctly")
	}

	// Verify series conversion
	if len(study.Series) != 1 {
		t.Fatalf("Expected 1 series")
	}
	series := study.Series[0]
	if series.Description != "Liver" {
		t.Errorf("Series description not converted correctly")
	}
	if series.CustomTags["BodyPartExamined"] != "LIVER" {
		t.Errorf("Series custom tags not converted correctly")
	}
}

func TestWizardStateToConfig(t *testing.T) {
	state := &WizardState{
		Global: types.GlobalConfig{
			Modality:          "XA",
			TotalImages:       200,
			TotalSize:         "2GB",
			OutputDir:         "/angio/output",
			Seed:              555,
			NumPatients:       10,
			StudiesPerPatient: 2,
			SeriesPerStudy:    5,
		},
		Patients: []types.PatientConfig{
			{
				Name:      "Angio Patient",
				ID:        "AP001",
				BirthDate: "1970-12-25",
				Sex:       "M",
				Studies: []types.StudyConfig{
					{
						Description:        "Coronary Angiography",
						Date:               "2024-04-01",
						AccessionNumber:    "XA001",
						Institution:        "Cardiac Center",
						Department:         "Cath Lab",
						BodyPart:           "HEART",
						Priority:           "STAT",
						ReferringPhysician: "Dr. Cardio",
						CustomTags: map[string]string{
							"ProtocolName": "CORONARY",
						},
						Series: []types.SeriesConfig{
							{
								Description: "LAO Cranial",
								Protocol:    "LAO_CRAN",
								Orientation: "LAO30_CRAN20",
								ImageCount:  100,
								CustomTags: map[string]string{
									"ViewPosition": "LAO",
								},
							},
						},
					},
				},
			},
		},
	}

	cfg := wizardStateToConfig(state)

	// Verify global conversion
	if cfg.Global.Modality != state.Global.Modality {
		t.Errorf("Modality not converted correctly")
	}
	if cfg.Global.TotalImages != state.Global.TotalImages {
		t.Errorf("TotalImages not converted correctly")
	}
	if cfg.Global.TotalSize != state.Global.TotalSize {
		t.Errorf("TotalSize not converted correctly")
	}
	if cfg.Global.OutputDir != state.Global.OutputDir {
		t.Errorf("OutputDir not converted correctly")
	}
	if cfg.Global.Seed != state.Global.Seed {
		t.Errorf("Seed not converted correctly")
	}
	if cfg.Global.NumPatients != state.Global.NumPatients {
		t.Errorf("NumPatients not converted correctly")
	}
	if cfg.Global.StudiesPerPatient != state.Global.StudiesPerPatient {
		t.Errorf("StudiesPerPatient not converted correctly")
	}
	if cfg.Global.SeriesPerStudy != state.Global.SeriesPerStudy {
		t.Errorf("SeriesPerStudy not converted correctly")
	}

	// Verify patient conversion
	if len(cfg.Patients) != 1 {
		t.Fatalf("Expected 1 patient")
	}
	if cfg.Patients[0].Name != "Angio Patient" {
		t.Errorf("Patient name not converted correctly")
	}

	// Verify study conversion
	if len(cfg.Patients[0].Studies) != 1 {
		t.Fatalf("Expected 1 study")
	}
	study := cfg.Patients[0].Studies[0]
	if study.Description != "Coronary Angiography" {
		t.Errorf("Study description not converted correctly")
	}
	if study.CustomTags["ProtocolName"] != "CORONARY" {
		t.Errorf("Study custom tags not converted correctly")
	}

	// Verify series conversion
	if len(study.Series) != 1 {
		t.Fatalf("Expected 1 series")
	}
	series := study.Series[0]
	if series.Description != "LAO Cranial" {
		t.Errorf("Series description not converted correctly")
	}
	if series.CustomTags["ViewPosition"] != "LAO" {
		t.Errorf("Series custom tags not converted correctly")
	}
}

func TestRoundtrip_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "roundtrip.yaml")

	original := &WizardState{
		Global: types.GlobalConfig{
			Modality:          "PT",
			TotalImages:       500,
			TotalSize:         "5GB",
			OutputDir:         "/pet/output",
			Seed:              777,
			NumPatients:       20,
			StudiesPerPatient: 1,
			SeriesPerStudy:    2,
		},
		Patients: []types.PatientConfig{
			{
				Name:      "PET Patient One",
				ID:        "PP001",
				BirthDate: "1960-05-15",
				Sex:       "F",
				Studies: []types.StudyConfig{
					{
						Description:        "Whole Body PET",
						Date:               "2024-05-01",
						AccessionNumber:    "PT001",
						Institution:        "Nuclear Medicine",
						Department:         "PET/CT",
						BodyPart:           "WHOLEBODY",
						Priority:           "ROUTINE",
						ReferringPhysician: "Dr. Nuclear",
						CustomTags: map[string]string{
							"RadiopharmaceuticalInformationSequence": "FDG",
						},
						Series: []types.SeriesConfig{
							{
								Description: "PET AC",
								Protocol:    "WB_PET_AC",
								Orientation: "TRANSVERSE",
								ImageCount:  300,
								CustomTags: map[string]string{
									"CorrectedImage": "ATTN",
								},
							},
							{
								Description: "PET NAC",
								Protocol:    "WB_PET_NAC",
								Orientation: "TRANSVERSE",
								ImageCount:  300,
								CustomTags: map[string]string{
									"CorrectedImage": "NONE",
								},
							},
						},
					},
				},
			},
			{
				Name:      "PET Patient Two",
				ID:        "PP002",
				BirthDate: "1975-10-30",
				Sex:       "M",
				Studies: []types.StudyConfig{
					{
						Description: "Brain PET",
						Date:        "2024-05-02",
						BodyPart:    "HEAD",
						Series: []types.SeriesConfig{
							{
								Description: "Brain PET",
								ImageCount:  100,
							},
						},
					},
				},
			},
		},
	}

	// Save
	if err := SaveToYAML(original, configPath); err != nil {
		t.Fatalf("SaveToYAML failed: %v", err)
	}

	// Load
	loaded, err := LoadFromYAML(configPath)
	if err != nil {
		t.Fatalf("LoadFromYAML failed: %v", err)
	}

	// Deep comparison of Global config
	if !reflect.DeepEqual(original.Global, loaded.Global) {
		t.Errorf("Global config mismatch:\nOriginal: %+v\nLoaded: %+v", original.Global, loaded.Global)
	}

	// Compare patient count
	if len(original.Patients) != len(loaded.Patients) {
		t.Fatalf("Patient count mismatch: original=%d, loaded=%d", len(original.Patients), len(loaded.Patients))
	}

	// Compare each patient
	for i, origPatient := range original.Patients {
		loadedPatient := loaded.Patients[i]

		if origPatient.Name != loadedPatient.Name {
			t.Errorf("Patient %d name mismatch", i)
		}
		if origPatient.ID != loadedPatient.ID {
			t.Errorf("Patient %d ID mismatch", i)
		}
		if origPatient.BirthDate != loadedPatient.BirthDate {
			t.Errorf("Patient %d birth date mismatch", i)
		}
		if origPatient.Sex != loadedPatient.Sex {
			t.Errorf("Patient %d sex mismatch", i)
		}

		// Compare studies
		if len(origPatient.Studies) != len(loadedPatient.Studies) {
			t.Fatalf("Patient %d study count mismatch", i)
		}

		for j, origStudy := range origPatient.Studies {
			loadedStudy := loadedPatient.Studies[j]

			if origStudy.Description != loadedStudy.Description {
				t.Errorf("Patient %d Study %d description mismatch", i, j)
			}
			if origStudy.BodyPart != loadedStudy.BodyPart {
				t.Errorf("Patient %d Study %d body part mismatch", i, j)
			}

			// Compare custom tags
			if !reflect.DeepEqual(origStudy.CustomTags, loadedStudy.CustomTags) {
				t.Errorf("Patient %d Study %d custom tags mismatch", i, j)
			}

			// Compare series
			if len(origStudy.Series) != len(loadedStudy.Series) {
				t.Fatalf("Patient %d Study %d series count mismatch", i, j)
			}

			for k, origSeries := range origStudy.Series {
				loadedSeries := loadedStudy.Series[k]

				if origSeries.Description != loadedSeries.Description {
					t.Errorf("Patient %d Study %d Series %d description mismatch", i, j, k)
				}
				if origSeries.ImageCount != loadedSeries.ImageCount {
					t.Errorf("Patient %d Study %d Series %d image count mismatch", i, j, k)
				}

				// Compare series custom tags
				if !reflect.DeepEqual(origSeries.CustomTags, loadedSeries.CustomTags) {
					t.Errorf("Patient %d Study %d Series %d custom tags mismatch", i, j, k)
				}
			}
		}
	}
}

func TestCopyMap(t *testing.T) {
	// Test nil map
	result := copyMap(nil)
	if result != nil {
		t.Errorf("Expected nil for nil input, got %v", result)
	}

	// Test empty map
	empty := make(map[string]string)
	result = copyMap(empty)
	if result == nil || len(result) != 0 {
		t.Errorf("Expected empty map for empty input")
	}

	// Test map with values
	original := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	result = copyMap(original)

	// Verify values are copied
	if result["key1"] != "value1" || result["key2"] != "value2" {
		t.Errorf("Values not copied correctly")
	}

	// Verify it's a new map (modify original, result should be unchanged)
	original["key1"] = "modified"
	if result["key1"] == "modified" {
		t.Errorf("copyMap did not create a true copy")
	}
}

func TestLoadFromYAML_MinimalConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "minimal.yaml")

	// Minimal valid config - just global section
	content := `
global:
  modality: CR
  total_images: 5
  total_size: 50MB
  output: ./minimal
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	state, err := LoadFromYAML(configPath)
	if err != nil {
		t.Fatalf("LoadFromYAML failed for minimal config: %v", err)
	}

	if state.Global.Modality != "CR" {
		t.Errorf("Expected modality CR, got %s", state.Global.Modality)
	}
	if state.Global.TotalImages != 5 {
		t.Errorf("Expected total_images 5, got %d", state.Global.TotalImages)
	}

	// Patients should be empty/nil
	if len(state.Patients) != 0 {
		t.Errorf("Expected 0 patients for minimal config, got %d", len(state.Patients))
	}
}

func TestSaveToYAML_InvalidPath(t *testing.T) {
	state := &WizardState{
		Global: types.GlobalConfig{
			Modality:    "MR",
			TotalImages: 10,
		},
	}

	// Try to save to an invalid path
	err := SaveToYAML(state, "/nonexistent/deeply/nested/path/config.yaml")
	if err == nil {
		t.Error("Expected error when saving to invalid path, got nil")
	}
}
