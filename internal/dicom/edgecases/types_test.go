package edgecases

import (
	"testing"
)

func TestParseTypes_Valid(t *testing.T) {
	types, err := ParseTypes("special-chars,long-names")
	if err != nil {
		t.Fatalf("ParseTypes failed: %v", err)
	}
	if len(types) != 2 {
		t.Errorf("Expected 2 types, got %d", len(types))
	}
	if types[0] != SpecialChars {
		t.Errorf("Expected SpecialChars, got %v", types[0])
	}
}

func TestParseTypes_Invalid(t *testing.T) {
	_, err := ParseTypes("invalid-type")
	if err == nil {
		t.Error("Expected error for invalid type")
	}
}

func TestParseTypes_All(t *testing.T) {
	types, err := ParseTypes("special-chars,long-names,missing-tags,old-dates,varied-ids")
	if err != nil {
		t.Fatalf("ParseTypes failed: %v", err)
	}
	if len(types) != 5 {
		t.Errorf("Expected 5 types, got %d", len(types))
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{"valid", Config{Percentage: 50, Types: []EdgeCaseType{SpecialChars}}, false},
		{"zero percent", Config{Percentage: 0, Types: []EdgeCaseType{SpecialChars}}, false},
		{"negative percent", Config{Percentage: -1, Types: []EdgeCaseType{SpecialChars}}, true},
		{"over 100 percent", Config{Percentage: 101, Types: []EdgeCaseType{SpecialChars}}, true},
		{"empty types with percent", Config{Percentage: 50, Types: []EdgeCaseType{}}, true},
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
	if (&Config{Percentage: 0}).IsEnabled() {
		t.Error("0% should not be enabled")
	}
	if (&Config{Percentage: 50, Types: []EdgeCaseType{}}).IsEnabled() {
		t.Error("Empty types should not be enabled")
	}
	if !(&Config{Percentage: 50, Types: []EdgeCaseType{SpecialChars}}).IsEnabled() {
		t.Error("50% with types should be enabled")
	}
}
