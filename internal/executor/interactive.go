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

// ExecuteInteractive runs a script in an interactive shell session.
// If tmux is available, the command runs in a new tmux window (if already
// inside tmux) or a new tmux session (if not). Otherwise, the command runs
// directly in the current terminal with stdin/stdout/stderr connected.
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

	logger.LogDebug("Created env file: %s", envFilePath)

	// Send pre-execution notification if configured
	if script.NotifyBefore != "" {
		sendNotification(script.Name, script.NotifyBefore, notification.NotificationContext{
			ScriptName: script.Name,
			Command:    command,
		})
	}

	// Check if tmux is available
	tmuxPath, tmuxErr := exec.LookPath("tmux")
	if tmuxErr == nil {
		logger.LogInfo("tmux found at %s, launching interactive session", tmuxPath)
		return executeInTmux(tmuxPath, script, command, clipboardContent, envFilePath)
	}

	// Fallback: run directly with connected stdio
	logger.LogInfo("tmux not available, running interactively in current terminal")
	return executeDirectInteractive(script, command, clipboardContent, envFilePath)
}

// executeInTmux launches the command in a tmux window or session.
// If already inside tmux ($TMUX is set), it creates a new window.
// Otherwise, it creates a new session and attaches to it.
func executeInTmux(tmuxPath string, script config.Script, command, clipboardContent, envFilePath string) error {
	// Build a wrapper script that sets environment and runs the command
	wrapperPath, err := buildInteractiveWrapper(command, clipboardContent, envFilePath)
	if err != nil {
		return fmt.Errorf("failed to create wrapper script: %w", err)
	}

	name := sanitizeTmuxName(script.Name)

	if os.Getenv("TMUX") != "" {
		// Inside tmux: create a new window
		logger.LogDebug("Inside tmux, creating new window: %s", name)
		cmd := exec.Command(tmuxPath, "new-window", "-n", name, wrapperPath)
		if err := cmd.Run(); err != nil {
			// Clean up wrapper on failure
			os.Remove(wrapperPath)
			os.Remove(envFilePath)
			return fmt.Errorf("failed to create tmux window: %w", err)
		}
		logger.LogInfo("Interactive session started in tmux window: %s", name)
		return nil
	}

	// Not in tmux: create a new session (attaches automatically)
	logger.LogDebug("Not inside tmux, creating new session: %s", name)
	cmd := exec.Command(tmuxPath, "new-session", "-s", name, wrapperPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		os.Remove(wrapperPath)
		os.Remove(envFilePath)
		return fmt.Errorf("failed to create tmux session: %w", err)
	}

	logger.LogInfo("Interactive tmux session ended: %s", name)
	return nil
}

// executeDirectInteractive runs the command directly with stdin/stdout/stderr
// connected to the current terminal, allowing the user to interact with it.
func executeDirectInteractive(script config.Script, command, clipboardContent, envFilePath string) error {
	defer os.Remove(envFilePath)

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

// buildInteractiveWrapper creates a temporary shell script that sets up the
// environment (CLIP, CLIPROUTER_ENV_FILE) and executes the command. The wrapper
// cleans up its own temp files on exit.
func buildInteractiveWrapper(command, clipboardContent, envFilePath string) (string, error) {
	wrapper, err := os.CreateTemp("", "cliprouter-interactive-*.sh")
	if err != nil {
		return "", err
	}
	wrapperPath := wrapper.Name()

	var content strings.Builder
	content.WriteString("#!/bin/sh\n")
	content.WriteString(fmt.Sprintf("export CLIP=%s\n", shellQuote(clipboardContent)))
	content.WriteString(fmt.Sprintf("export CLIPROUTER_ENV_FILE=%s\n", shellQuote(envFilePath)))

	if hasShebang(command) {
		// Write the shebang script to a separate temp file so the
		// interpreter specified by the shebang is used.
		scriptFile, err := os.CreateTemp("", "cliprouter-cmd-*")
		if err != nil {
			wrapper.Close()
			os.Remove(wrapperPath)
			return "", err
		}
		scriptPath := scriptFile.Name()
		if _, err := scriptFile.WriteString(command); err != nil {
			scriptFile.Close()
			os.Remove(scriptPath)
			wrapper.Close()
			os.Remove(wrapperPath)
			return "", err
		}
		scriptFile.Close()
		if err := os.Chmod(scriptPath, 0700); err != nil {
			os.Remove(scriptPath)
			wrapper.Close()
			os.Remove(wrapperPath)
			return "", err
		}

		content.WriteString(fmt.Sprintf("%s\n", shellQuote(scriptPath)))
		// Clean up the script file after execution
		content.WriteString(fmt.Sprintf("rm -f %s\n", shellQuote(scriptPath)))
	} else {
		content.WriteString(command)
		content.WriteString("\n")
	}

	// Self-cleanup of wrapper and env file
	content.WriteString(fmt.Sprintf("rm -f %s %s\n", shellQuote(wrapperPath), shellQuote(envFilePath)))

	if _, err := wrapper.WriteString(content.String()); err != nil {
		wrapper.Close()
		os.Remove(wrapperPath)
		return "", err
	}
	wrapper.Close()

	if err := os.Chmod(wrapperPath, 0700); err != nil {
		os.Remove(wrapperPath)
		return "", err
	}

	return wrapperPath, nil
}

// shellQuote returns a shell-safe single-quoted string.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// sanitizeTmuxName converts a script name to a valid tmux session/window name.
func sanitizeTmuxName(name string) string {
	name = strings.ReplaceAll(name, ".", "-")
	name = strings.ReplaceAll(name, ":", "-")
	name = strings.ReplaceAll(name, " ", "-")
	if len(name) > 30 {
		name = name[:30]
	}
	return name
}
