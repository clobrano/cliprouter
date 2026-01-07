package executor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/clobrano/cliprouter/internal/config"
	"github.com/clobrano/cliprouter/internal/logger"
)

// Execute runs a script with the clipboard content
func Execute(script config.Script, clipboardContent string, timeout time.Duration) (string, string, error) {
	logger.LogInfo("Executing script: %s", script.Name)

	// Substitute ${CLIP} placeholder
	command := substituteClipboard(script.Command, clipboardContent)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Detect if multiline (shell script) or simple command
	var cmd *exec.Cmd
	if isMultilineCommand(script.Command) {
		// Execute as shell script
		cmd = exec.CommandContext(ctx, "/bin/sh", "-c", command)
		logger.LogDebug("Executing as shell script: /bin/sh -c %s", command)
	} else {
		// Parse and execute as simple command
		cmd = createSimpleCommand(ctx, command)
		logger.LogDebug("Executing as simple command: %v", cmd.Args)
	}

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	err := cmd.Run()

	stdoutStr := stdout.String()
	stderrStr := stderr.String()

	// Log output
	if stdoutStr != "" {
		logger.LogDebug("stdout: %s", stdoutStr)
	}
	if stderrStr != "" {
		logger.LogDebug("stderr: %s", stderrStr)
	}

	// Check for timeout
	if ctx.Err() == context.DeadlineExceeded {
		logger.LogError("Script '%s' timed out after %v", script.Name, timeout)
		return stdoutStr, stderrStr, fmt.Errorf("script execution timed out after %v", timeout)
	}

	if err != nil {
		logger.LogError("Script '%s' failed: %v", script.Name, err)
		return stdoutStr, stderrStr, fmt.Errorf("script execution failed: %w", err)
	}

	logger.LogInfo("Script '%s' completed successfully", script.Name)
	return stdoutStr, stderrStr, nil
}

// substituteClipboard replaces ${CLIP} with the clipboard content
func substituteClipboard(command, clipboardContent string) string {
	// Escape the clipboard content for shell safety
	escaped := escapeShellString(clipboardContent)
	return strings.ReplaceAll(command, "${CLIP}", escaped)
}

// escapeShellString escapes a string for safe use in shell commands
func escapeShellString(s string) string {
	// Wrap in single quotes and escape any existing single quotes
	s = strings.ReplaceAll(s, "'", "'\"'\"'")
	return "'" + s + "'"
}

// isMultilineCommand checks if the command contains newlines
func isMultilineCommand(command string) bool {
	return strings.Contains(command, "\n")
}

// createSimpleCommand parses a simple command into exec.Command
func createSimpleCommand(ctx context.Context, command string) *exec.Cmd {
	// Simple parsing: split by spaces, but this is basic
	// For production, consider using shell parsing library
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return exec.CommandContext(ctx, "true") // no-op
	}

	if len(parts) == 1 {
		return exec.CommandContext(ctx, parts[0])
	}

	return exec.CommandContext(ctx, parts[0], parts[1:]...)
}

// SubstituteClipboard is a public wrapper for substituteClipboard (used in dry-run)
func SubstituteClipboard(command, clipboardContent string) string {
	return substituteClipboard(command, clipboardContent)
}
