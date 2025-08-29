package parser

import "testing"

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"single word", "user", "User"},
		{"snake case", "user_name", "UserName"},
		{"multiple underscores", "user_full_name_data", "UserFullNameData"},
		{"leading underscore", "_user_name", "UserName"},
		{"trailing underscore", "user_name_", "UserName"},
		{"consecutive underscores", "user__name", "UserName"},
		{"single character", "a", "A"},
		{"mixed case input", "User_Name", "UserName"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toCamelCase(tt.input)
			if result != tt.expected {
				t.Errorf("toCamelCase(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCapitalizeFirst(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"lowercase word", "hello", "Hello"},
		{"uppercase word", "HELLO", "HELLO"},
		{"mixed case", "hELLO", "HELLO"},
		{"single character", "a", "A"},
		{"single uppercase", "A", "A"},
		{"unicode", "ñame", "Ñame"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := capitalizeFirst(tt.input)
			if result != tt.expected {
				t.Errorf("capitalizeFirst(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}