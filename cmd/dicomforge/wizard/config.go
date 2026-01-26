package wizard

import (
	"fmt"
	"os"

	"github.com/mrsinham/dicomforge/cmd/dicomforge/wizard/types"
	"gopkg.in/yaml.v3"
)

// Config represents the complete wizard configuration for YAML serialization.
type Config struct {
	Global   GlobalConfigYAML    `yaml:"global"`
	Patients []PatientConfigYAML `yaml:"patients"`
}

// LoadFromYAML reads a config file and returns WizardState.
func LoadFromYAML(path string) (*WizardState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	return configToWizardState(&cfg), nil
}

// SaveToYAML writes WizardState to a YAML file.
func SaveToYAML(state *WizardState, path string) error {
	cfg := wizardStateToConfig(state)

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// configToWizardState converts Config (YAML) to WizardState (runtime).
func configToWizardState(c *Config) *WizardState {
	state := &WizardState{
		Global: types.GlobalConfig{
			Modality:          c.Global.Modality,
			TotalImages:       c.Global.TotalImages,
			TotalSize:         c.Global.TotalSize,
			OutputDir:         c.Global.OutputDir,
			Seed:              c.Global.Seed,
			NumPatients:       c.Global.NumPatients,
			StudiesPerPatient: c.Global.StudiesPerPatient,
			SeriesPerStudy:    c.Global.SeriesPerStudy,
		},
		Patients: make([]types.PatientConfig, len(c.Patients)),
	}

	for i, p := range c.Patients {
		patient := types.PatientConfig{
			Name:      p.Name,
			ID:        p.ID,
			BirthDate: p.BirthDate,
			Sex:       p.Sex,
			Studies:   make([]types.StudyConfig, len(p.Studies)),
		}

		for j, s := range p.Studies {
			study := types.StudyConfig{
				Description:        s.Description,
				Date:               s.Date,
				AccessionNumber:    s.AccessionNumber,
				Institution:        s.Institution,
				Department:         s.Department,
				BodyPart:           s.BodyPart,
				Priority:           s.Priority,
				ReferringPhysician: s.ReferringPhysician,
				CustomTags:         copyMap(s.CustomTags),
				Series:             make([]types.SeriesConfig, len(s.Series)),
			}

			for k, ser := range s.Series {
				study.Series[k] = types.SeriesConfig{
					Description: ser.Description,
					Protocol:    ser.Protocol,
					Orientation: ser.Orientation,
					ImageCount:  ser.ImageCount,
					CustomTags:  copyMap(ser.CustomTags),
				}
			}

			patient.Studies[j] = study
		}

		state.Patients[i] = patient
	}

	return state
}

// wizardStateToConfig converts WizardState to Config (for YAML serialization).
func wizardStateToConfig(s *WizardState) *Config {
	cfg := &Config{
		Global: GlobalConfigYAML{
			Modality:          s.Global.Modality,
			TotalImages:       s.Global.TotalImages,
			TotalSize:         s.Global.TotalSize,
			OutputDir:         s.Global.OutputDir,
			Seed:              s.Global.Seed,
			NumPatients:       s.Global.NumPatients,
			StudiesPerPatient: s.Global.StudiesPerPatient,
			SeriesPerStudy:    s.Global.SeriesPerStudy,
		},
		Patients: make([]PatientConfigYAML, len(s.Patients)),
	}

	for i, p := range s.Patients {
		patient := PatientConfigYAML{
			Name:      p.Name,
			ID:        p.ID,
			BirthDate: p.BirthDate,
			Sex:       p.Sex,
			Studies:   make([]StudyConfigYAML, len(p.Studies)),
		}

		for j, st := range p.Studies {
			study := StudyConfigYAML{
				Description:        st.Description,
				Date:               st.Date,
				AccessionNumber:    st.AccessionNumber,
				Institution:        st.Institution,
				Department:         st.Department,
				BodyPart:           st.BodyPart,
				Priority:           st.Priority,
				ReferringPhysician: st.ReferringPhysician,
				CustomTags:         copyMap(st.CustomTags),
				Series:             make([]SeriesConfigYAML, len(st.Series)),
			}

			for k, ser := range st.Series {
				study.Series[k] = SeriesConfigYAML{
					Description: ser.Description,
					Protocol:    ser.Protocol,
					Orientation: ser.Orientation,
					ImageCount:  ser.ImageCount,
					CustomTags:  copyMap(ser.CustomTags),
				}
			}

			patient.Studies[j] = study
		}

		cfg.Patients[i] = patient
	}

	return cfg
}

// copyMap creates a copy of a string map.
func copyMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// GlobalConfigYAML holds global settings with YAML tags for serialization.
type GlobalConfigYAML struct {
	Modality          string `yaml:"modality"`
	TotalImages       int    `yaml:"total_images"`
	TotalSize         string `yaml:"total_size"`
	OutputDir         string `yaml:"output"`
	Seed              int64  `yaml:"seed,omitempty"`
	NumPatients       int    `yaml:"num_patients,omitempty"`
	StudiesPerPatient int    `yaml:"studies_per_patient,omitempty"`
	SeriesPerStudy    int    `yaml:"series_per_study,omitempty"`
}

// PatientConfigYAML holds patient configuration with YAML tags.
type PatientConfigYAML struct {
	Name      string            `yaml:"name"`
	ID        string            `yaml:"id"`
	BirthDate string            `yaml:"birth_date"`
	Sex       string            `yaml:"sex"`
	Studies   []StudyConfigYAML `yaml:"studies"`
}

// StudyConfigYAML holds study configuration with YAML tags.
type StudyConfigYAML struct {
	Description        string             `yaml:"description"`
	Date               string             `yaml:"date"`
	AccessionNumber    string             `yaml:"accession"`
	Institution        string             `yaml:"institution"`
	Department         string             `yaml:"department"`
	BodyPart           string             `yaml:"body_part"`
	Priority           string             `yaml:"priority"`
	ReferringPhysician string             `yaml:"referring_physician"`
	CustomTags         map[string]string  `yaml:"custom_tags,omitempty"`
	Series             []SeriesConfigYAML `yaml:"series"`
}

// SeriesConfigYAML holds series configuration with YAML tags.
type SeriesConfigYAML struct {
	Description string            `yaml:"description"`
	Protocol    string            `yaml:"protocol"`
	Orientation string            `yaml:"orientation"`
	ImageCount  int               `yaml:"images"`
	CustomTags  map[string]string `yaml:"custom_tags,omitempty"`
}
