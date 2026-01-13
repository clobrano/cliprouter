// +build windows

package notification

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/clobrano/cliprouter/internal/logger"
)

// sendPlatformNotification sends a notification on Windows using PowerShell
func sendPlatformNotification(title, message string) error {
	// Build PowerShell command to create a Windows notification
	// Using Windows.UI.Notifications.ToastNotificationManager

	var displayTitle, displayMessage string
	if title != "" {
		displayTitle = title
	} else {
		displayTitle = "cliprouter"
	}

	if message != "" {
		displayMessage = message
	} else {
		displayMessage = displayTitle
		displayTitle = "cliprouter"
	}

	// Escape single quotes for PowerShell
	safeTitle := escapePowerShell(displayTitle)
	safeMessage := escapePowerShell(displayMessage)

	// PowerShell script to show notification
	// Using a simpler approach with Add-Type and System.Windows.Forms
	script := fmt.Sprintf(`
[void] [System.Reflection.Assembly]::LoadWithPartialName("System.Windows.Forms")
$notification = New-Object System.Windows.Forms.NotifyIcon
$notification.Icon = [System.Drawing.SystemIcons]::Information
$notification.BalloonTipIcon = [System.Windows.Forms.ToolTipIcon]::Info
$notification.BalloonTipText = '%s'
$notification.BalloonTipTitle = '%s'
$notification.Visible = $true
$notification.ShowBalloonTip(5000)
Start-Sleep -Seconds 1
$notification.Dispose()
`, safeMessage, safeTitle)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)

	logger.LogDebug("Executing PowerShell notification script")

	if err := cmd.Run(); err != nil {
		logger.LogError("Failed to send notification: %v", err)
		return fmt.Errorf("failed to send notification: %w", err)
	}

	logger.LogDebug("Notification sent successfully")
	return nil
}

// escapePowerShell escapes special characters for PowerShell strings
func escapePowerShell(s string) string {
	// Escape single quotes by doubling them
	s = strings.ReplaceAll(s, "'", "''")
	// Escape backticks
	s = strings.ReplaceAll(s, "`", "``")
	return s
}

// sendPlatformActionNotification shows an interactive dialog on Windows using PowerShell MessageBox
// Returns true if user clicked Yes/OK, false if No/Cancel
func sendPlatformActionNotification(title, prompt string) (bool, error) {
	// Escape strings for PowerShell
	safeTitle := escapePowerShell(title)
	safePrompt := escapePowerShell(prompt)

	// PowerShell script to show a MessageBox with Yes/No buttons
	// MessageBoxButtons.YesNo = 4
	// MessageBoxIcon.Question = 32
	// DialogResult.Yes = 6
	script := fmt.Sprintf(`
[void] [System.Reflection.Assembly]::LoadWithPartialName("System.Windows.Forms")
$result = [System.Windows.Forms.MessageBox]::Show('%s', '%s', 'YesNo', 'Question')
if ($result -eq 'Yes') {
    exit 0
} else {
    exit 1
}
`, safePrompt, safeTitle)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	logger.LogDebug("Executing PowerShell dialog script")

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			logger.LogDebug("PowerShell dialog exited with code %d", exitCode)
			// Exit code 0 = Yes, 1 = No
			return exitCode == 0, nil
		}
		return false, fmt.Errorf("failed to show dialog: %w", err)
	}

	logger.LogDebug("User confirmed action")
	return true, nil
}
