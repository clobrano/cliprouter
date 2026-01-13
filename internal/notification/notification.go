package notification

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

const (
	defaultTitle = "cliprouter"
)

// Send sends a desktop notification with the given message
// The title defaults to "cliprouter" but can be customized
func Send(message string) error {
	return SendWithTitle(defaultTitle, message)
}

// SendWithTitle sends a desktop notification with a custom title and message
func SendWithTitle(title, message string) error {
	if message == "" {
		return nil // No message, nothing to send
	}

	switch runtime.GOOS {
	case "linux":
		return sendLinux(title, message)
	case "darwin":
		return sendMacOS(title, message)
	case "windows":
		return sendWindows(title, message)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// sendLinux sends a notification on Linux using notify-send
func sendLinux(title, message string) error {
	cmd := exec.Command("notify-send", title, message)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send notification (is notify-send installed?): %w", err)
	}
	return nil
}

// sendMacOS sends a notification on macOS using osascript
func sendMacOS(title, message string) error {
	// Escape quotes in the message for AppleScript
	escapedMessage := strings.ReplaceAll(message, `"`, `\"`)
	escapedTitle := strings.ReplaceAll(title, `"`, `\"`)

	script := fmt.Sprintf(`display notification "%s" with title "%s"`, escapedMessage, escapedTitle)
	cmd := exec.Command("osascript", "-e", script)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	return nil
}

// sendWindows sends a notification on Windows using PowerShell
func sendWindows(title, message string) error {
	// Escape quotes for PowerShell
	escapedMessage := strings.ReplaceAll(message, `"`, `'`)
	escapedTitle := strings.ReplaceAll(title, `"`, `'`)

	// Use PowerShell to show a toast notification
	script := fmt.Sprintf(
		`[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] > $null; `+
		`$template = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02); `+
		`$xml = [xml]$template.GetXml(); `+
		`$xml.toast.visual.binding.text[0].InnerText = "%s"; `+
		`$xml.toast.visual.binding.text[1].InnerText = "%s"; `+
		`$toast = [Windows.UI.Notifications.ToastNotification]::new($xml); `+
		`[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("cliprouter").Show($toast)`,
		escapedTitle, escapedMessage)

	cmd := exec.Command("powershell", "-Command", script)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	return nil
}

// SubstituteVariables replaces placeholders in the notification message
// Supported variables:
// - ${ERROR}: The error message
// - ${CLIP}: The clipboard content
func SubstituteVariables(message string, clipContent string, errorMsg string) string {
	result := message
	result = strings.ReplaceAll(result, "${CLIP}", clipContent)
	result = strings.ReplaceAll(result, "${ERROR}", errorMsg)
	return result
}
