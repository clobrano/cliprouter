package executor

import (
	"os"
	"strings"
	"testing"

	"github.com/clobrano/cliprouter/internal/config"
)

func TestShellQuote(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple string", "hello", "'hello'"},
		{"with spaces", "hello world", "'hello world'"},
		{"with single quote", "don't", "'don'\\''t'"},
		{"empty string", "", "''"},
		{"with special chars", "foo;bar&&baz", "'foo;bar&&baz'"},
		{"with newlines", "line1\nline2", "'line1\nline2'"},
		{"with double quotes", `say "hi"`, `'say "hi"'`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shellQuote(tt.input)
			if result != tt.expected {
				t.Errorf("shellQuote(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeTmuxName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple name", "test", "test"},
		{"with spaces", "My Script", "My-Script"},
		{"with dots", "version.1.0", "version-1-0"},
		{"with colons", "task:build", "task-build"},
		{"long name truncated", "this is a very long script name that exceeds thirty characters", "this-is-a-very-long-script-nam"},
		{"mixed special chars", "Build: v1.2", "Build--v1-2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeTmuxName(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeTmuxName(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuildInteractiveWrapper(t *testing.T) {
	t.Run("simple command", func(t *testing.T) {
		wrapperPath, err := buildInteractiveWrapper("echo hello", "clipboard content", "/tmp/test-env")
		if err != nil {
			t.Fatalf("buildInteractiveWrapper() error = %v", err)
		}
		defer os.Remove(wrapperPath)

		content, err := os.ReadFile(wrapperPath)
		if err != nil {
			t.Fatalf("Failed to read wrapper: %v", err)
		}

		contentStr := string(content)

		// Verify shebang
		if !strings.HasPrefix(contentStr, "#!/bin/sh\n") {
			t.Error("Wrapper should start with #!/bin/sh")
		}

		// Verify CLIP export
		if !strings.Contains(contentStr, "export CLIP='clipboard content'") {
			t.Error("Wrapper should export CLIP variable")
		}

		// Verify CLIPROUTER_ENV_FILE export
		if !strings.Contains(contentStr, "export CLIPROUTER_ENV_FILE='/tmp/test-env'") {
			t.Error("Wrapper should export CLIPROUTER_ENV_FILE variable")
		}

		// Verify command is included
		if !strings.Contains(contentStr, "echo hello") {
			t.Error("Wrapper should contain the command")
		}

		// Verify self-cleanup
		if !strings.Contains(contentStr, "rm -f") {
			t.Error("Wrapper should contain cleanup commands")
		}

		// Verify executable permissions
		info, err := os.Stat(wrapperPath)
		if err != nil {
			t.Fatalf("Failed to stat wrapper: %v", err)
		}
		if info.Mode()&0700 != 0700 {
			t.Errorf("Wrapper should be executable, got mode %v", info.Mode())
		}
	})

	t.Run("command with single quotes in clipboard", func(t *testing.T) {
		wrapperPath, err := buildInteractiveWrapper("echo ${CLIP}", "don't stop", "/tmp/test-env")
		if err != nil {
			t.Fatalf("buildInteractiveWrapper() error = %v", err)
		}
		defer os.Remove(wrapperPath)

		content, err := os.ReadFile(wrapperPath)
		if err != nil {
			t.Fatalf("Failed to read wrapper: %v", err)
		}

		// Single quotes in clipboard content should be properly escaped
		if !strings.Contains(string(content), "'don'\\''t stop'") {
			t.Error("Wrapper should properly escape single quotes in CLIP")
		}
	})

	t.Run("shebang command", func(t *testing.T) {
		command := "#!/usr/bin/env bash\necho hello"
		wrapperPath, err := buildInteractiveWrapper(command, "test", "/tmp/test-env")
		if err != nil {
			t.Fatalf("buildInteractiveWrapper() error = %v", err)
		}
		defer os.Remove(wrapperPath)

		content, err := os.ReadFile(wrapperPath)
		if err != nil {
			t.Fatalf("Failed to read wrapper: %v", err)
		}

		contentStr := string(content)

		// Shebang scripts should be written to a separate file and referenced
		// The wrapper should NOT contain the shebang directly
		if strings.Contains(contentStr, "#!/usr/bin/env bash") {
			t.Error("Shebang command should be in a separate script file, not inline in wrapper")
		}

		// Should reference a temp script file
		if !strings.Contains(contentStr, "/tmp/cliprouter-cmd-") {
			t.Error("Wrapper should reference a separate temp script file for shebang commands")
		}
	})
}

func TestExecuteDirectInteractive(t *testing.T) {
	t.Run("simple command completes", func(t *testing.T) {
		script := config.Script{
			Name:    "Test Interactive",
			Command: "true",
		}

		err := executeDirectInteractive(script, script.Command, "test", createTempEnvFile(t))
		if err != nil {
			t.Errorf("executeDirectInteractive() error = %v", err)
		}
	})

	t.Run("failing command returns error", func(t *testing.T) {
		script := config.Script{
			Name:    "Test Fail",
			Command: "false",
		}

		err := executeDirectInteractive(script, script.Command, "test", createTempEnvFile(t))
		if err == nil {
			t.Error("executeDirectInteractive() should have returned error for failing command")
		}
	})

	t.Run("command receives CLIP env var", func(t *testing.T) {
		// Create a temp file to capture the env var value
		outFile, err := os.CreateTemp("", "cliprouter-test-out-*")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		outPath := outFile.Name()
		outFile.Close()
		defer os.Remove(outPath)

		script := config.Script{
			Name:    "Test Env",
			Command: "echo ${CLIP} > " + outPath,
		}

		err = executeDirectInteractive(script, script.Command, "test-clipboard-value", createTempEnvFile(t))
		if err != nil {
			t.Errorf("executeDirectInteractive() error = %v", err)
		}

		content, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("Failed to read output: %v", err)
		}

		if !strings.Contains(string(content), "test-clipboard-value") {
			t.Errorf("Command should receive CLIP env var, got output: %q", string(content))
		}
	})
}

// createTempEnvFile creates a temporary env file for testing and registers cleanup.
func createTempEnvFile(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("", "cliprouter-test-env-*")
	if err != nil {
		t.Fatalf("Failed to create temp env file: %v", err)
	}
	path := f.Name()
	f.Close()
	t.Cleanup(func() { os.Remove(path) })
	return path
}
