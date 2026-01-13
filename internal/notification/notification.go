package notification

import (
	"fmt"
	"os"
	"strings"

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
			// Try to get from environment
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
