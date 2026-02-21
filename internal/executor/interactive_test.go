package executor

import (
	"os"
	"strings"
	"testing"

	"github.com/clobrano/cliprouter/internal/config"
)

func TestExecuteDirectInteractive(t *testing.T) {
	t.Run("simple command completes", func(t *testing.T) {
		script := config.Script{
			Name:    "Test Interactive",
			Command: "true",
		}

		err := ExecuteInteractive(script, "test")
		if err != nil {
			t.Errorf("ExecuteInteractive() error = %v", err)
		}
	})

	t.Run("failing command returns error", func(t *testing.T) {
		script := config.Script{
			Name:    "Test Fail",
			Command: "false",
		}

		err := ExecuteInteractive(script, "test")
		if err == nil {
			t.Error("ExecuteInteractive() should have returned error for failing command")
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

		err = ExecuteInteractive(script, "test-clipboard-value")
		if err != nil {
			t.Errorf("ExecuteInteractive() error = %v", err)
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
