package wizard

import (
	"github.com/mrsinham/dicomforge/cmd/dicomforge/wizard/types"
	"github.com/mrsinham/dicomforge/internal/dicom"
	"github.com/mrsinham/dicomforge/internal/dicom/modalities"
	"github.com/mrsinham/dicomforge/internal/util"
)

// ToGeneratorOptions converts WizardState to GeneratorOptions for generation.
func ToGeneratorOptions(s *WizardState) (dicom.GeneratorOptions, error) {
	// Calculate total images from series
	totalImages := 0
	totalStudies := 0
	for _, p := range s.Patients {
		for _, st := range p.Studies {
			totalStudies++
			for _, ser := range st.Series {
				totalImages += ser.ImageCount
			}
		}
	}

	// If no detailed patients, use global config
	if len(s.Patients) == 0 {
		totalImages = s.Global.TotalImages
		totalStudies = s.Global.NumPatients * s.Global.StudiesPerPatient
	}

	// Build PredefinedPatients from WizardState
	predefined := make([]dicom.PredefinedPatient, len(s.Patients))
	for i, p := range s.Patients {
		patient := dicom.PredefinedPatient{
			Name:      p.Name,
			ID:        p.ID,
			BirthDate: p.BirthDate,
			Sex:       p.Sex,
			Studies:   make([]dicom.PredefinedStudy, len(p.Studies)),
		}

		for j, st := range p.Studies {
			study := dicom.PredefinedStudy{
				Description:        st.Description,
				Date:               st.Date,
				AccessionNumber:    st.AccessionNumber,
				Institution:        st.Institution,
				Department:         st.Department,
				BodyPart:           st.BodyPart,
				Priority:           st.Priority,
				ReferringPhysician: st.ReferringPhysician,
				Series:             make([]dicom.PredefinedSeries, len(st.Series)),
			}

			for k, ser := range st.Series {
				study.Series[k] = dicom.PredefinedSeries{
					Description: ser.Description,
					Protocol:    ser.Protocol,
					Orientation: ser.Orientation,
					ImageCount:  ser.ImageCount,
				}
			}

			patient.Studies[j] = study
		}

		predefined[i] = patient
	}

	// Parse modality
	mod := modalities.Modality(s.Global.Modality)
	if s.Global.Modality == "" {
		mod = modalities.MR
	}

	// Build series range
	seriesPerStudy := util.SeriesRange{Min: s.Global.SeriesPerStudy, Max: s.Global.SeriesPerStudy}
	if seriesPerStudy.Min == 0 {
		seriesPerStudy = util.SeriesRange{Min: 1, Max: 1}
	}

	return dicom.GeneratorOptions{
		NumImages:          totalImages,
		TotalSize:          s.Global.TotalSize,
		OutputDir:          s.Global.OutputDir,
		Seed:               s.Global.Seed,
		NumStudies:         totalStudies,
		NumPatients:        len(s.Patients),
		Modality:           mod,
		SeriesPerStudy:     seriesPerStudy,
		PredefinedPatients: predefined,
	}, nil
}

// FromGeneratorOptions creates a WizardState from GeneratorOptions.
// Used for --save-config to export CLI options as YAML.
func FromGeneratorOptions(opts dicom.GeneratorOptions) *WizardState {
	numPatients := opts.NumPatients
	if numPatients == 0 {
		numPatients = 1
	}

	studiesPerPatient := 1
	if opts.NumStudies > 0 && numPatients > 0 {
		studiesPerPatient = opts.NumStudies / numPatients
	}

	seriesPerStudy := opts.SeriesPerStudy.Min
	if seriesPerStudy == 0 {
		seriesPerStudy = 1
	}

	state := &WizardState{
		Global: types.GlobalConfig{
			Modality:          string(opts.Modality),
			TotalImages:       opts.NumImages,
			TotalSize:         opts.TotalSize,
			OutputDir:         opts.OutputDir,
			Seed:              opts.Seed,
			NumPatients:       numPatients,
			StudiesPerPatient: studiesPerPatient,
			SeriesPerStudy:    seriesPerStudy,
		},
	}

	// If PredefinedPatients exist, convert them back
	if len(opts.PredefinedPatients) > 0 {
		state.Patients = make([]types.PatientConfig, len(opts.PredefinedPatients))
		for i, p := range opts.PredefinedPatients {
			patient := types.PatientConfig{
				Name:      p.Name,
				ID:        p.ID,
				BirthDate: p.BirthDate,
				Sex:       p.Sex,
				Studies:   make([]types.StudyConfig, len(p.Studies)),
			}

			for j, st := range p.Studies {
				study := types.StudyConfig{
					Description:        st.Description,
					Date:               st.Date,
					AccessionNumber:    st.AccessionNumber,
					Institution:        st.Institution,
					Department:         st.Department,
					BodyPart:           st.BodyPart,
					Priority:           st.Priority,
					ReferringPhysician: st.ReferringPhysician,
					Series:             make([]types.SeriesConfig, len(st.Series)),
				}

				for k, ser := range st.Series {
					study.Series[k] = types.SeriesConfig{
						Description: ser.Description,
						Protocol:    ser.Protocol,
						Orientation: ser.Orientation,
						ImageCount:  ser.ImageCount,
					}
				}

				patient.Studies[j] = study
			}

			state.Patients[i] = patient
		}
	}

	return state
}
