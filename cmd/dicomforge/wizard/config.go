package wizard

// Config represents the complete wizard configuration for YAML serialization.
type Config struct {
	Global   GlobalConfigYAML   `yaml:"global"`
	Patients []PatientConfigYAML `yaml:"patients"`
}

// GlobalConfigYAML holds global settings with YAML tags for serialization.
type GlobalConfigYAML struct {
	Modality          string `yaml:"modality"`
	TotalImages       int    `yaml:"total_images"`
	TotalSize         string `yaml:"total_size"`
	OutputDir         string `yaml:"output_dir"`
	Seed              int64  `yaml:"seed"`
	NumPatients       int    `yaml:"num_patients"`
	StudiesPerPatient int    `yaml:"studies_per_patient"`
	SeriesPerStudy    int    `yaml:"series_per_study"`
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
	AccessionNumber    string             `yaml:"accession_number"`
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
	ImageCount  int               `yaml:"image_count"`
	CustomTags  map[string]string `yaml:"custom_tags,omitempty"`
}
