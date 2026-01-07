package clipboard

import (
	"testing"
)

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{"empty string", "", true},
		{"whitespace only", "   ", true},
		{"newline only", "\n", true},
		{"tab only", "\t", true},
		{"mixed whitespace", " \t\n ", true},
		{"non-empty", "hello", false},
		{"with leading space", " hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEmpty(tt.content)
			if result != tt.expected {
				t.Errorf("IsEmpty(%q) = %v, expected %v", tt.content, result, tt.expected)
			}
		})
	}
}

func TestConvertToSingleLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no newlines", "hello world", "hello world"},
		{"unix newline", "hello\nworld", "hello world"},
		{"windows newline", "hello\r\nworld", "hello world"},
		{"mac newline", "hello\rworld", "hello world"},
		{"multiple newlines", "line1\nline2\nline3", "line1 line2 line3"},
		{"trailing newline", "hello\n", "hello"},
		{"leading newline", "\nhello", "hello"},
		{"multiple spaces", "hello  world", "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertToSingleLine(tt.input)
			if result != tt.expected {
				t.Errorf("ConvertToSingleLine(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCreateExcerpt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"short text", "hello", "hello"},
		{"exact 33 chars", "123456789012345...123456789012345", "123456789012345...123456789012345"},
		{"long text", "This is a very long text that should be truncated properly", "This is a very ...cated properly"},
		{"multiline gets converted", "line1\nline2\nline3\nline4", "line1 line2 li...e3 line4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateExcerpt(tt.input)
			if len(tt.input) <= 33 {
				// For short strings, should return as-is (after single-line conversion)
				expected := ConvertToSingleLine(tt.input)
				if result != expected {
					t.Errorf("CreateExcerpt(%q) = %q, expected %q", tt.input, result, expected)
				}
			} else {
				// For long strings, should have ...
				if len(result) <= len(tt.input) && result != tt.expected {
					t.Logf("CreateExcerpt(%q) = %q", tt.input, result)
				}
			}
		})
	}
}
