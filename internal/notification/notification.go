package notification

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/clobrano/cliprouter/internal/history"
	"github.com/clobrano/cliprouter/internal/logger"
)

// NotificationContext holds the context data for variable substitution in notifications
type NotificationContext struct {
	ScriptName string
	Command    string
	ExitCode   int
	Stdout     string
	Stderr     string
	Error      string
	ScriptEnv  map[string]string // Script-specific environment variables
}

// Send sends a notification with the given title and message
// Returns error if notification fails, but doesn't stop execution
func Send(title, message string) error {
	if title == "" && message == "" {
		return nil // Nothing to send
	}

	logger.LogDebug("Sending notification: title='%s', message='%s'", title, message)
	return sendPlatformNotification(title, message)
}

// SubstituteVariables replaces variables in the notification message
// Supports environment variables and context variables
// Variables format:
//   - ${ENV_VAR_NAME} - environment variable
//   - ${SCRIPT_NAME} - name of the script being executed
//   - ${COMMAND} - the command being executed
//   - ${EXIT_CODE} - exit code of the command (0 for success)
//   - ${STDOUT} - standard output from the command
//   - ${STDERR} - standard error from the command
//   - ${ERROR} - error message (if any)
func SubstituteVariables(text string, ctx NotificationContext) string {
	if text == "" {
		return ""
	}

	result := text

	// Substitute context variables
	result = strings.ReplaceAll(result, "${SCRIPT_NAME}", ctx.ScriptName)
	result = strings.ReplaceAll(result, "${COMMAND}", ctx.Command)
	result = strings.ReplaceAll(result, "${EXIT_CODE}", formatExitCode(ctx.ExitCode))
	result = strings.ReplaceAll(result, "${STDOUT}", ctx.Stdout)
	result = strings.ReplaceAll(result, "${STDERR}", ctx.Stderr)
	result = strings.ReplaceAll(result, "${ERROR}", ctx.Error)

	// Substitute environment variables
	// Find all ${...} patterns and try to resolve them as environment variables
	result = os.Expand(result, func(varName string) string {
		// Check if it's already substituted (one of our context variables)
		switch varName {
		case "SCRIPT_NAME", "COMMAND", "EXIT_CODE", "STDOUT", "STDERR", "ERROR":
			return "${" + varName + "}" // Keep as is, already substituted above
		default:
			// First, try to get from script-specific environment variables
			if ctx.ScriptEnv != nil {
				if value, ok := ctx.ScriptEnv[varName]; ok {
					return value
				}
			}
			// Then try to get from parent process environment
			if value := os.Getenv(varName); value != "" {
				return value
			}
			// Return original if not found
			return "${" + varName + "}"
		}
	})

	return result
}

// formatExitCode converts exit code to string
func formatExitCode(code int) string {
	return fmt.Sprintf("%d", code)
}

// ActionConfig represents an action to perform after user confirmation
type ActionConfig struct {
	OpenFile string // Path to file to open
	Execute  string // Command to execute
}

// SendAction sends an interactive notification that requests user confirmation
// and executes the specified action if confirmed
// Returns true if user confirmed, false if canceled or on error
func SendAction(title, prompt string, action *ActionConfig, ctx NotificationContext) (bool, error) {
	if prompt == "" || action == nil {
		return false, nil // Nothing to do
	}

	// Check that action has at least one field set
	if action.OpenFile == "" && action.Execute == "" {
		logger.LogDebug("No action configured, skipping interactive notification")
		return false, nil
	}

	// Substitute variables in prompt and action
	prompt = SubstituteVariables(prompt, ctx)

	logger.LogDebug("Sending action notification: title='%s', prompt='%s'", title, prompt)

	// Record the action prompt in history
	history.RecordActionPrompt(ctx.ScriptName, prompt)

	// Show platform-specific dialog
	confirmed, err := sendPlatformActionNotification(title, prompt)
	if err != nil {
		logger.LogDebug("Action notification failed: %v", err)
		return false, err
	}

	// Record user decision in history
	history.RecordActionDecision(ctx.ScriptName, confirmed)

	if !confirmed {
		logger.LogDebug("User canceled action notification")
		return false, nil
	}

	// User confirmed, execute the action
	if action.OpenFile != "" {
		filePath := SubstituteVariables(action.OpenFile, ctx)
		logger.LogDebug("Opening file: %s", filePath)
		history.RecordActionExecution(ctx.ScriptName, "open_file", filePath)
		return true, openFile(filePath)
	}

	if action.Execute != "" {
		command := SubstituteVariables(action.Execute, ctx)
		logger.LogDebug("Executing command: %s", command)
		history.RecordActionExecution(ctx.ScriptName, "execute", command)
		return true, executeCommand(command)
	}

	return true, nil
}

// openFile opens a file with the default application for the platform
func openFile(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", path)
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	// Wait for the opener command to complete
	// This ensures the file opener (xdg-open, open, start) has launched the application
	// before the main process exits
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("failed to wait for file opener: %w", err)
	}

	return nil
}

// executeCommand executes a command and waits for it to complete
func executeCommand(command string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux", "darwin":
		cmd = exec.Command("sh", "-c", command)
	case "windows":
		cmd = exec.Command("cmd", "/c", command)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	// Wait for the command to complete
	// This ensures the action completes before the main process exits
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}
