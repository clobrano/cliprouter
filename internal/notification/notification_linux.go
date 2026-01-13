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
