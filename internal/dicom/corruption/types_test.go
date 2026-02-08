package corruption

import (
	"testing"
)

func TestParseTypes(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []CorruptionType
		wantErr bool
	}{
		{
			name:  "empty",
			input: "",
			want:  nil,
		},
		{
			name:  "single type",
			input: "siemens-csa",
			want:  []CorruptionType{SiemensCSA},
		},
		{
			name:  "multiple types",
			input: "siemens-csa,ge-private",
			want:  []CorruptionType{SiemensCSA, GEPrivate},
		},
		{
			name:  "all types",
			input: "all",
			want:  AllCorruptionTypes(),
		},
		{
			name:  "with whitespace",
			input: " siemens-csa , ge-private ",
			want:  []CorruptionType{SiemensCSA, GEPrivate},
		},
		{
			name:    "invalid type",
			input:   "invalid-type",
			wantErr: true,
		},
		{
			name:  "all corruption types individually",
			input: "siemens-csa,ge-private,philips-private,malformed-lengths",
			want:  []CorruptionType{SiemensCSA, GEPrivate, PhilipsPrivate, MalformedLengths},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTypes(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTypes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("ParseTypes() got %d types, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ParseTypes()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "no types",
			config:  Config{},
			wantErr: true,
		},
		{
			name: "valid single type",
			config: Config{
				Types: []CorruptionType{SiemensCSA},
			},
		},
		{
			name: "valid all types",
			config: Config{
				Types: AllCorruptionTypes(),
			},
		},
		{
			name: "invalid type",
			config: Config{
				Types: []CorruptionType{"invalid"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_IsEnabled(t *testing.T) {
	empty := Config{}
	if empty.IsEnabled() {
		t.Error("empty config should not be enabled")
	}

	enabled := Config{Types: []CorruptionType{SiemensCSA}}
	if !enabled.IsEnabled() {
		t.Error("config with types should be enabled")
	}
}

func TestConfig_HasType(t *testing.T) {
	config := Config{Types: []CorruptionType{SiemensCSA, GEPrivate}}

	if !config.HasType(SiemensCSA) {
		t.Error("should have SiemensCSA")
	}
	if !config.HasType(GEPrivate) {
		t.Error("should have GEPrivate")
	}
	if config.HasType(PhilipsPrivate) {
		t.Error("should not have PhilipsPrivate")
	}
	if config.HasType(MalformedLengths) {
		t.Error("should not have MalformedLengths")
	}
}
