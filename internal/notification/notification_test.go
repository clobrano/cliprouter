package notification

import (
	"testing"
)

func TestSubstituteVariables(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		clipContent string
		errorMsg    string
		expected    string
	}{
		{
			name:        "substitute clip content",
			message:     "Processing: ${CLIP}",
			clipContent: "hello world",
			errorMsg:    "",
			expected:    "Processing: hello world",
		},
		{
			name:        "substitute error message",
			message:     "Error occurred: ${ERROR}",
			clipContent: "",
			errorMsg:    "command failed",
			expected:    "Error occurred: command failed",
		},
		{
			name:        "substitute both clip and error",
			message:     "Failed to process ${CLIP}: ${ERROR}",
			clipContent: "test data",
			errorMsg:    "timeout",
			expected:    "Failed to process test data: timeout",
		},
		{
			name:        "no substitution needed",
			message:     "Simple message",
			clipContent: "ignored",
			errorMsg:    "ignored",
			expected:    "Simple message",
		},
		{
			name:        "empty message",
			message:     "",
			clipContent: "test",
			errorMsg:    "error",
			expected:    "",
		},
		{
			name:        "multiple clip substitutions",
			message:     "${CLIP} and ${CLIP} again",
			clipContent: "data",
			errorMsg:    "",
			expected:    "data and data again",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SubstituteVariables(tt.message, tt.clipContent, tt.errorMsg)
			if result != tt.expected {
				t.Errorf("SubstituteVariables() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSend(t *testing.T) {
	// Test sending with empty message (should not error)
	err := Send("")
	if err != nil {
		t.Errorf("Send(\"\") should not error, got: %v", err)
	}

	// Note: We can't easily test actual notification sending without mocking
	// the system commands, but we can at least verify the function doesn't panic
	// and handles empty messages correctly
}

func TestSendWithTitle(t *testing.T) {
	// Test sending with empty message (should not error)
	err := SendWithTitle("Test Title", "")
	if err != nil {
		t.Errorf("SendWithTitle() with empty message should not error, got: %v", err)
	}

	// Note: We can't easily test actual notification sending without mocking
	// the system commands, but we can at least verify the function doesn't panic
	// and handles empty messages correctly
}
