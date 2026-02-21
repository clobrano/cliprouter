package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/clobrano/cliprouter/internal/config"
	"github.com/clobrano/cliprouter/internal/logger"
	"github.com/clobrano/cliprouter/internal/notification"
)

// ExecuteInteractive runs a script in the current terminal with stdin/stdout/stderr
// connected, allowing the user to interact with it directly.
func ExecuteInteractive(script config.Script, clipboardContent string) error {
	logger.LogInfo("Executing interactive script: %s", script.Name)

	command := script.Command

	// Create a temporary file for dynamic environment variables
	envFile, err := os.CreateTemp("", "cliprouter-env-*.txt")
	if err != nil {
		logger.LogError("Failed to create env file: %v", err)
		return fmt.Errorf("failed to create env file: %w", err)
	}
	envFilePath := envFile.Name()
	envFile.Close()
	defer os.Remove(envFilePath)

	logger.LogDebug("Created env file: %s", envFilePath)

	// Send pre-execution notification if configured
	if script.NotifyBefore != "" {
		sendNotification(script.Name, script.NotifyBefore, notification.NotificationContext{
			ScriptName: script.Name,
			Command:    command,
		})
	}

	var cmd *exec.Cmd

	if isMultilineCommand(command) || strings.Contains(command, "${CLIP}") {
		if hasShebang(command) {
			// Write script to temp file to respect shebang
			scriptFile, err := os.CreateTemp("", "cliprouter-script-*")
			if err != nil {
				return fmt.Errorf("failed to create script file: %w", err)
			}
			scriptPath := scriptFile.Name()
			defer os.Remove(scriptPath)

			if _, err := scriptFile.WriteString(command); err != nil {
				scriptFile.Close()
				return fmt.Errorf("failed to write script file: %w", err)
			}
			scriptFile.Close()

			if err := os.Chmod(scriptPath, 0700); err != nil {
				return fmt.Errorf("failed to make script executable: %w", err)
			}

			cmd = exec.Command(scriptPath)
			logger.LogDebug("Executing interactive shebang script: %s", scriptPath)
		} else {
			shell := os.Getenv("SHELL")
			if shell == "" {
				shell = "/bin/sh"
			}
			cmd = exec.Command(shell, "-c", command)
			logger.LogDebug("Executing interactive shell script: %s -c ...", shell)
		}
	} else {
		cmd = createSimpleCommand(context.Background(), command)
		logger.LogDebug("Executing interactive simple command: %v", cmd.Args)
	}

	// Set environment
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("CLIP=%s", clipboardContent))
	cmd.Env = append(cmd.Env, fmt.Sprintf("CLIPROUTER_ENV_FILE=%s", envFilePath))

	// Connect stdio for interactive use
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.LogDebug("Starting interactive process for script: %s", script.Name)
	if err := cmd.Run(); err != nil {
		logger.LogError("Interactive script '%s' failed: %v", script.Name, err)
		return fmt.Errorf("interactive script execution failed: %w", err)
	}

	logger.LogInfo("Interactive script '%s' completed", script.Name)
	return nil
}
