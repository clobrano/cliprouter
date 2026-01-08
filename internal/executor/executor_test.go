package executor

import (
	"strings"
	"testing"
	"time"

	"github.com/clobrano/cliprouter/internal/config"
)

func TestSubstituteClipboard(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		clipboard string
		contains  string // Check if output contains this
	}{
		{
			name:      "simple substitution",
			command:   "echo ${CLIP}",
			clipboard: "hello",
			contains:  "hello",
		},
		{
			name:      "multiple substitutions",
			command:   "echo ${CLIP} and ${CLIP}",
			clipboard: "test",
			contains:  "test",
		},
		{
			name:      "no substitution",
			command:   "echo hello",
			clipboard: "world",
			contains:  "echo hello",
		},
		{
			name:      "special characters",
			command:   "echo ${CLIP}",
			clipboard: "hello'world",
			contains:  "hello", // Should be escaped
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := substituteClipboard(tt.command, tt.clipboard)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("substituteClipboard() = %q, should contain %q", result, tt.contains)
			}
		})
	}
}

func TestEscapeShellString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "hello", "'hello'"},
		{"with spaces", "hello world", "'hello world'"},
		{"with single quote", "it's", "'it'\"'\"'s'"},
		{"empty", "", "''"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeShellString(tt.input)
			if result != tt.expected {
				t.Errorf("escapeShellString(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsMultilineCommand(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{"single line", "echo hello", false},
		{"multiline", "echo hello\necho world", true},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isMultilineCommand(tt.command)
			if result != tt.expected {
				t.Errorf("isMultilineCommand(%q) = %v, expected %v", tt.command, result, tt.expected)
			}
		})
	}
}

func TestExecute(t *testing.T) {
	tests := []struct {
		name        string
		script      config.Script
		clipboard   string
		expectError bool
	}{
		{
			name: "simple echo",
			script: config.Script{
				Name:    "Echo Test",
				Command: "echo ${CLIP}",
			},
			clipboard:   "hello",
			expectError: false,
		},
		{
			name: "invalid command",
			script: config.Script{
				Name:    "Invalid",
				Command: "nonexistentcommand12345",
			},
			clipboard:   "test",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeout := 5 * time.Second
			_, _, err := Execute(tt.script, tt.clipboard, timeout)
			if (err != nil) != tt.expectError {
				t.Errorf("Execute() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestExecuteTimeout(t *testing.T) {
	script := config.Script{
		Name:    "Sleep Test",
		Command: "sleep 5",
	}

	timeout := 1 * time.Second
	_, _, err := Execute(script, "test", timeout)

	if err == nil {
		t.Error("Execute() should have timed out but didn't")
	}

	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

// TestExecuteWaitsForCompletion verifies that Execute waits for the process to complete
func TestExecuteWaitsForCompletion(t *testing.T) {
	// Use multiline to ensure it's executed as a shell script
	script := config.Script{
		Name: "Wait Test",
		Command: `sleep 1
echo 'completed'`,
	}

	timeout := 5 * time.Second
	start := time.Now()
	stdout, _, err := Execute(script, "test", timeout)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Execute() should not have failed: %v", err)
	}

	// Verify that at least 1 second elapsed (process was waited for)
	if elapsed < 1*time.Second {
		t.Errorf("Execute() returned too quickly (%v), should have waited for process", elapsed)
	}

	// Verify that the output contains "completed", indicating process finished
	if !strings.Contains(stdout, "completed") {
		t.Errorf("Execute() stdout = %q, should contain 'completed'", stdout)
	}
}

// TestExecuteWaitsForShellScript verifies waiting for multi-line shell scripts
func TestExecuteWaitsForShellScript(t *testing.T) {
	script := config.Script{
		Name: "Multi-line Wait Test",
		Command: `
sleep 1
echo 'step1'
sleep 1
echo 'step2'
`,
	}

	timeout := 10 * time.Second
	start := time.Now()
	stdout, _, err := Execute(script, "test", timeout)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Execute() should not have failed: %v", err)
	}

	// Verify that at least 2 seconds elapsed (process was waited for)
	if elapsed < 2*time.Second {
		t.Errorf("Execute() returned too quickly (%v), should have waited at least 2 seconds", elapsed)
	}

	// Verify both steps completed
	if !strings.Contains(stdout, "step1") || !strings.Contains(stdout, "step2") {
		t.Errorf("Execute() stdout = %q, should contain both 'step1' and 'step2'", stdout)
	}
}
