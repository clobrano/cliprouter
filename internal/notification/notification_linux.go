// +build linux

package notification

import (
	"fmt"
	"os/exec"

	"github.com/clobrano/cliprouter/internal/logger"
)

// sendPlatformNotification sends a notification on Linux using notify-send
func sendPlatformNotification(title, message string) error {
	// Check if notify-send is available
	_, err := exec.LookPath("notify-send")
	if err != nil {
		logger.LogDebug("notify-send not found, skipping notification")
		return fmt.Errorf("notify-send not available: %w", err)
	}

	// Build the command
	// notify-send [options] <summary> [body]
	args := []string{}

	// Add urgency level (normal by default)
	args = append(args, "-u", "normal")

	// Add app name
	args = append(args, "-a", "cliprouter")

	// Add title and message
	if title != "" && message != "" {
		args = append(args, title, message)
	} else if title != "" {
		args = append(args, title)
	} else if message != "" {
		args = append(args, "cliprouter", message)
	}

	cmd := exec.Command("notify-send", args...)

	logger.LogDebug("Executing: notify-send %v", args)

	if err := cmd.Run(); err != nil {
		logger.LogError("Failed to send notification: %v", err)
		return fmt.Errorf("failed to send notification: %w", err)
	}

	logger.LogDebug("Notification sent successfully")
	return nil
}

// sendPlatformActionNotification shows an interactive dialog on Linux using zenity
// Returns true if user clicked Yes/OK, false if No/Cancel or on error
func sendPlatformActionNotification(title, prompt string) (bool, error) {
	// Try zenity first (most common)
	if _, err := exec.LookPath("zenity"); err == nil {
		return sendZenityDialog(title, prompt)
	}

	// Try kdialog as fallback for KDE
	if _, err := exec.LookPath("kdialog"); err == nil {
		return sendKdialog(title, prompt)
	}

	logger.LogDebug("No dialog tool found (zenity or kdialog), cannot show interactive notification")
	return false, fmt.Errorf("no dialog tool available (zenity or kdialog)")
}

// sendZenityDialog shows a question dialog using zenity
func sendZenityDialog(title, prompt string) (bool, error) {
	args := []string{
		"--question",
		"--title=" + title,
		"--text=" + prompt,
		"--no-wrap",
	}

	cmd := exec.Command("zenity", args...)
	logger.LogDebug("Executing: zenity %v", args)

	// zenity returns exit code 0 for Yes, 1 for No, 5 for timeout
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			logger.LogDebug("zenity exited with code %d", exitCode)
			return exitCode == 0, nil // 0 = Yes, anything else = No/Cancel
		}
		return false, fmt.Errorf("failed to run zenity: %w", err)
	}

	logger.LogDebug("User confirmed action")
	return true, nil
}

// sendKdialog shows a question dialog using kdialog
func sendKdialog(title, prompt string) (bool, error) {
	args := []string{
		"--title", title,
		"--yesno", prompt,
	}

	cmd := exec.Command("kdialog", args...)
	logger.LogDebug("Executing: kdialog %v", args)

	// kdialog returns exit code 0 for Yes, 1 for No
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			logger.LogDebug("kdialog exited with code %d", exitCode)
			return exitCode == 0, nil
		}
		return false, fmt.Errorf("failed to run kdialog: %w", err)
	}

	logger.LogDebug("User confirmed action")
	return true, nil
}
