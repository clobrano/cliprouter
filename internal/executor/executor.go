package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
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

	// Detect if multiline (shell script) or simple command
	// Force shell execution if the command contains ${CLIP} to allow shell variable expansion
	var cmd *exec.Cmd
	if isMultilineCommand(script.Command) || strings.Contains(script.Command, "${CLIP}") {
		// Check if script has a shebang - if so, write to temp file to respect it
		if hasShebang(script.Command) {
			scriptFile, err := os.CreateTemp("", "cliprouter-script-*")
			if err != nil {
				logger.LogError("Failed to create script file: %v", err)
				return "", "", fmt.Errorf("failed to create script file: %w", err)
			}
			scriptPath := scriptFile.Name()
			defer os.Remove(scriptPath)

			if _, err := scriptFile.WriteString(command); err != nil {
				scriptFile.Close()
				logger.LogError("Failed to write script file: %v", err)
				return "", "", fmt.Errorf("failed to write script file: %w", err)
			}
			scriptFile.Close()

			if err := os.Chmod(scriptPath, 0700); err != nil {
				logger.LogError("Failed to make script executable: %v", err)
				return "", "", fmt.Errorf("failed to make script executable: %w", err)
			}

			cmd = exec.Command(scriptPath)
			logger.LogDebug("Executing script file with shebang: %s", scriptPath)
		} else {
			shell := os.Getenv("SHELL")
			if shell == "" {
				shell = "/bin/sh"
			}
			cmd = exec.Command(shell, "-c", command)
			logger.LogDebug("Executing as shell script: %s -c %s", shell, command)
		}
	} else {
		cmd = createSimpleCommand(context.Background(), command)
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

	// Detach from the parent process group so the command survives cliprouter exiting
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	// Start the command and return immediately (background execution)
	logger.LogDebug("Starting background process for script: %s", script.Name)
	if err = cmd.Start(); err != nil {
		logger.LogError("Failed to start script '%s': %v", script.Name, err)
		return "", "", fmt.Errorf("failed to start script: %w", err)
	}

	logger.LogInfo("Script '%s' started in background (pid %d)", script.Name, cmd.Process.Pid)
	return "", "", nil
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
