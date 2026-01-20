package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/clobrano/cliprouter/internal/config"
	"github.com/clobrano/cliprouter/internal/logger"
	"github.com/clobrano/cliprouter/internal/notification"
)

// Execute runs a script with the clipboard content
// If timeout is 0, the script will run without a timeout
func Execute(script config.Script, clipboardContent string, timeout time.Duration) (string, string, error) {
	logger.LogInfo("Executing script: %s", script.Name)

	// Use the command as-is; clipboard content will be passed via CLIP env var
	command := script.Command

	// Create a temporary file for dynamic environment variables
	envFile, err := os.CreateTemp("", "cliprouter-env-*.txt")
	if err != nil {
		logger.LogError("Failed to create env file: %v", err)
		return "", "", fmt.Errorf("failed to create env file: %w", err)
	}
	envFilePath := envFile.Name()
	envFile.Close()
	defer os.Remove(envFilePath) // Clean up after execution

	logger.LogDebug("Created env file: %s", envFilePath)

	// Send pre-execution notification if configured
	if script.NotifyBefore != "" {
		sendNotification(script.Name, script.NotifyBefore, notification.NotificationContext{
			ScriptName: script.Name,
			Command:    command,
			ExitCode:   0,
			Stdout:     "",
			Stderr:     "",
			Error:      "",
			ScriptEnv:  nil,
		})
	}

	// Create context with or without timeout
	var ctx context.Context
	var cancel context.CancelFunc

	if timeout > 0 {
		logger.LogDebug("Using timeout: %v", timeout)
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		logger.LogDebug("Running without timeout - will wait indefinitely")
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	// Detect if multiline (shell script) or simple command
	// Force shell execution if the command contains ${CLIP} to allow shell variable expansion
	var cmd *exec.Cmd
	if isMultilineCommand(script.Command) || strings.Contains(script.Command, "${CLIP}") {
		// Check if script has a shebang - if so, write to temp file to respect it
		if hasShebang(script.Command) {
			// Write script to temp file and execute it to respect shebang
			// No extension needed - the shebang specifies the interpreter
			scriptFile, err := os.CreateTemp("", "cliprouter-script-*")
			if err != nil {
				logger.LogError("Failed to create script file: %v", err)
				return "", "", fmt.Errorf("failed to create script file: %w", err)
			}
			scriptPath := scriptFile.Name()
			defer os.Remove(scriptPath)

			// Write the script content
			if _, err := scriptFile.WriteString(command); err != nil {
				scriptFile.Close()
				logger.LogError("Failed to write script file: %v", err)
				return "", "", fmt.Errorf("failed to write script file: %w", err)
			}
			scriptFile.Close()

			// Make it executable
			if err := os.Chmod(scriptPath, 0700); err != nil {
				logger.LogError("Failed to make script executable: %v", err)
				return "", "", fmt.Errorf("failed to make script executable: %w", err)
			}

			cmd = exec.CommandContext(ctx, scriptPath)
			logger.LogDebug("Executing script file with shebang: %s", scriptPath)
		} else {
			// Execute as shell script using user's shell for better environment support
			shell := os.Getenv("SHELL")
			if shell == "" {
				shell = "/bin/sh"
			}
			cmd = exec.CommandContext(ctx, shell, "-c", command)
			logger.LogDebug("Executing as shell script: %s -c %s", shell, command)
		}
	} else {
		// Parse and execute as simple command
		cmd = createSimpleCommand(ctx, command)
		logger.LogDebug("Executing as simple command: %v", cmd.Args)
	}

	// Set environment variables for the command
	// Always start with parent process environment
	cmd.Env = os.Environ()

	// Add the CLIP variable with clipboard content
	cmd.Env = append(cmd.Env, fmt.Sprintf("CLIP=%s", clipboardContent))
	logger.LogDebug("Setting env var: CLIP=[clipboard content]")

	// Add the CLIPROUTER_ENV_FILE variable
	cmd.Env = append(cmd.Env, fmt.Sprintf("CLIPROUTER_ENV_FILE=%s", envFilePath))
	logger.LogDebug("Setting env var: CLIPROUTER_ENV_FILE=%s", envFilePath)

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Start the command
	logger.LogDebug("Starting process for script: %s", script.Name)
	err = cmd.Start()
	if err != nil {
		logger.LogError("Failed to start script '%s': %v", script.Name, err)
		return "", "", fmt.Errorf("failed to start script: %w", err)
	}

	// Wait for the process to complete
	logger.LogDebug("Waiting for process to complete: %s", script.Name)
	err = cmd.Wait()

	stdoutStr := stdout.String()
	stderrStr := stderr.String()

	// Log output
	if stdoutStr != "" {
		logger.LogDebug("stdout: %s", stdoutStr)
	}
	if stderrStr != "" {
		logger.LogDebug("stderr: %s", stderrStr)
	}

	// Check for timeout (only relevant if timeout was set)
	if timeout > 0 && ctx.Err() == context.DeadlineExceeded {
		logger.LogError("Script '%s' timed out after %v", script.Name, timeout)
		return stdoutStr, stderrStr, fmt.Errorf("script execution timed out after %v", timeout)
	}

	// Determine exit code and error message
	exitCode := 0
	errorMsg := ""
	hasError := false
	if err != nil {
		exitCode = 1
		errorMsg = err.Error()
		hasError = true
		logger.LogError("Script '%s' failed: %v", script.Name, err)
	} else {
		logger.LogInfo("Script '%s' completed successfully - process has exited", script.Name)
	}

	// Read dynamic environment variables from the env file
	var finalEnv map[string]string
	if envVars, err := readEnvFile(envFilePath); err == nil {
		finalEnv = envVars
		for k, v := range envVars {
			logger.LogDebug("Dynamic env var from script: %s=%s", k, v)
		}
	} else {
		logger.LogDebug("No dynamic env vars or error reading env file: %v", err)
		finalEnv = make(map[string]string)
	}

	// Send post-execution notification based on success or failure
	notifCtx := notification.NotificationContext{
		ScriptName: script.Name,
		Command:    command,
		ExitCode:   exitCode,
		Stdout:     stdoutStr,
		Stderr:     stderrStr,
		Error:      errorMsg,
		ScriptEnv:  finalEnv,
	}

	if hasError && script.NotifyOnError != "" {
		// Send error notification if command failed
		sendNotification(script.Name, script.NotifyOnError, notifCtx)
	} else if !hasError && script.NotifyAfter != "" {
		// Send success notification if command succeeded
		sendNotification(script.Name, script.NotifyAfter, notifCtx)
	}

	// Send interactive action notification if configured (only on success)
	if !hasError && script.NotifyAction != nil {
		sendActionNotification(script.Name, script.NotifyAction, notifCtx)
	}

	if err != nil {
		return stdoutStr, stderrStr, fmt.Errorf("script execution failed: %w", err)
	}

	return stdoutStr, stderrStr, nil
}

// substituteClipboard replaces ${CLIP} with the clipboard content for display purposes
// Note: In actual execution, CLIP is passed as an environment variable
func substituteClipboard(command, clipboardContent string) string {
	return strings.ReplaceAll(command, "${CLIP}", clipboardContent)
}

// isMultilineCommand checks if the command contains newlines
func isMultilineCommand(command string) bool {
	return strings.Contains(command, "\n")
}

// hasShebang checks if the command starts with a shebang (#!)
func hasShebang(command string) bool {
	return strings.HasPrefix(strings.TrimSpace(command), "#!")
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

// sendNotification sends a notification with variable substitution
// The title is always the script name, only the message supports variable substitution
func sendNotification(title, message string, ctx notification.NotificationContext) {
	if message == "" {
		return
	}

	// Substitute variables in message
	message = notification.SubstituteVariables(message, ctx)

	// Send the notification
	if err := notification.Send(title, message); err != nil {
		// Log the error but don't fail the execution
		logger.LogError("Failed to send notification: %v", err)
	}
}

// sendActionNotification sends an interactive notification that executes an action on confirmation
func sendActionNotification(title string, notifyAction *config.NotifyAction, ctx notification.NotificationContext) {
	if notifyAction == nil || notifyAction.Prompt == "" || notifyAction.OnConfirm == nil {
		return
	}

	// Convert config.ActionConfig to notification.ActionConfig
	action := &notification.ActionConfig{
		OpenFile: notifyAction.OnConfirm.OpenFile,
		Execute:  notifyAction.OnConfirm.Execute,
	}

	// Send the action notification (this will show dialog and execute action if confirmed)
	confirmed, err := notification.SendAction(title, notifyAction.Prompt, action, ctx)
	if err != nil {
		// Log the error but don't fail the execution
		logger.LogError("Failed to send action notification: %v", err)
		return
	}

	if confirmed {
		logger.LogInfo("User confirmed action notification")
	} else {
		logger.LogInfo("User declined action notification")
	}
}

// readEnvFile reads environment variables from a file
// Expected format: VAR_NAME=value (one per line)
// Lines starting with # are ignored as comments
// Empty lines are ignored
func readEnvFile(path string) (map[string]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	envVars := make(map[string]string)
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse VAR_NAME=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key != "" {
				envVars[key] = value
			}
		}
	}

	return envVars, nil
}
