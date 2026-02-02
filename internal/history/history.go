package history

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"
)

var (
	historyFile *os.File
)

const HistoryFileName = "cliprouter/history.log"

// ActivityType represents the type of activity being recorded
type ActivityType string

const (
	ActivityNotifyBefore  ActivityType = "NOTIFY_BEFORE"
	ActivityNotifyAfter   ActivityType = "NOTIFY_AFTER"
	ActivityNotifyOnError ActivityType = "NOTIFY_ERROR"
	ActivityActionPrompt  ActivityType = "ACTION_PROMPT"
	ActivityActionConfirm ActivityType = "ACTION_CONFIRMED"
	ActivityActionDecline ActivityType = "ACTION_DECLINED"
	ActivityActionExecute ActivityType = "ACTION_EXECUTE"
	ActivityActionOpenFile ActivityType = "ACTION_OPEN_FILE"
)

// InitHistory initializes the history logging system
func InitHistory() error {
	// Get history file path
	historyPath := filepath.Join(xdg.DataHome, HistoryFileName)

	// Create directory if it doesn't exist
	dir := filepath.Dir(historyPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create history directory: %w", err)
	}

	// Open history file in append mode
	var err error
	historyFile, err = os.OpenFile(historyPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open history file: %w", err)
	}

	return nil
}

// Close closes the history file
func Close() {
	if historyFile != nil {
		historyFile.Close()
	}
}

// GetHistoryPath returns the path to the history file
func GetHistoryPath() string {
	return filepath.Join(xdg.DataHome, HistoryFileName)
}

// Record records an activity to the history file
func Record(activityType ActivityType, scriptName, message string) {
	if historyFile == nil {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	entry := fmt.Sprintf("[%s] [%s] script=%q %s\n", timestamp, activityType, scriptName, message)

	historyFile.WriteString(entry)
}

// RecordNotification records a notification activity
func RecordNotification(activityType ActivityType, scriptName, notificationMessage string) {
	Record(activityType, scriptName, fmt.Sprintf("message=%q", notificationMessage))
}

// RecordActionPrompt records when an action prompt is shown to the user
func RecordActionPrompt(scriptName, prompt string) {
	Record(ActivityActionPrompt, scriptName, fmt.Sprintf("prompt=%q", prompt))
}

// RecordActionDecision records the user's decision on an action prompt
func RecordActionDecision(scriptName string, confirmed bool) {
	if confirmed {
		Record(ActivityActionConfirm, scriptName, "user confirmed action")
	} else {
		Record(ActivityActionDecline, scriptName, "user declined action")
	}
}

// RecordActionExecution records the execution of an action after user confirmation
func RecordActionExecution(scriptName, actionType, actionValue string) {
	if actionType == "open_file" {
		Record(ActivityActionOpenFile, scriptName, fmt.Sprintf("path=%q", actionValue))
	} else if actionType == "execute" {
		Record(ActivityActionExecute, scriptName, fmt.Sprintf("command=%q", actionValue))
	}
}
