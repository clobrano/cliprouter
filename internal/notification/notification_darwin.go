// +build darwin

package notification

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/clobrano/cliprouter/internal/logger"
)

// sendPlatformNotification sends a notification on macOS using osascript
func sendPlatformNotification(title, message string) error {
	// Build AppleScript command
	// display notification "message" with title "title"
	var script string

	if title != "" && message != "" {
		// Escape quotes in title and message
		safeTitle := escapeAppleScript(title)
		safeMessage := escapeAppleScript(message)
		script = fmt.Sprintf(`display notification "%s" with title "%s"`, safeMessage, safeTitle)
	} else if title != "" {
		safeTitle := escapeAppleScript(title)
		script = fmt.Sprintf(`display notification "%s"`, safeTitle)
	} else if message != "" {
		safeMessage := escapeAppleScript(message)
		script = fmt.Sprintf(`display notification "%s"`, safeMessage)
	} else {
		return nil // Nothing to send
	}

	cmd := exec.Command("osascript", "-e", script)

	logger.LogDebug("Executing: osascript -e '%s'", script)

	if err := cmd.Run(); err != nil {
		logger.LogError("Failed to send notification: %v", err)
		return fmt.Errorf("failed to send notification: %w", err)
	}

	logger.LogDebug("Notification sent successfully")
	return nil
}

// escapeAppleScript escapes special characters for AppleScript strings
func escapeAppleScript(s string) string {
	// Escape backslashes first, then quotes
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// sendPlatformActionNotification shows an interactive dialog on macOS using AppleScript
// Returns true if user clicked Yes/OK, false if No/Cancel
func sendPlatformActionNotification(title, prompt string) (bool, error) {
	// Build AppleScript command for dialog
	// display dialog "prompt" with title "title" buttons {"No", "Yes"} default button "Yes"
	safeTitle := escapeAppleScript(title)
	safePrompt := escapeAppleScript(prompt)

	script := fmt.Sprintf(
		`display dialog "%s" with title "%s" buttons {"No", "Yes"} default button "Yes"`,
		safePrompt, safeTitle,
	)

	cmd := exec.Command("osascript", "-e", script)
	logger.LogDebug("Executing: osascript -e '%s'", script)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if user clicked No or Cancel
		if strings.Contains(string(output), "User canceled") {
			logger.LogDebug("User canceled action")
			return false, nil
		}
		return false, fmt.Errorf("failed to show dialog: %w", err)
	}

	// If no error, user clicked Yes
	// Output will be: "button returned:Yes"
	result := string(output)
	logger.LogDebug("Dialog result: %s", result)

	confirmed := strings.Contains(result, "Yes")
	if confirmed {
		logger.LogDebug("User confirmed action")
	} else {
		logger.LogDebug("User declined action")
	}

	return confirmed, nil
}
