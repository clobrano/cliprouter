package clipboard

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
)

// GetClipboard reads the current clipboard content
func GetClipboard() (string, error) {
	content, err := clipboard.ReadAll()
	if err != nil {
		return "", fmt.Errorf("failed to read clipboard (ensure xclip/xsel/wl-clipboard is installed on Linux): %w", err)
	}
	return content, nil
}

// IsEmpty checks if the clipboard content is empty
func IsEmpty(content string) bool {
	return strings.TrimSpace(content) == ""
}

// ConvertToSingleLine replaces all newlines with spaces
func ConvertToSingleLine(content string) string {
	// Replace various newline types with space
	content = strings.ReplaceAll(content, "\r\n", " ")
	content = strings.ReplaceAll(content, "\n", " ")
	content = strings.ReplaceAll(content, "\r", " ")

	// Collapse multiple spaces into one
	for strings.Contains(content, "  ") {
		content = strings.ReplaceAll(content, "  ", " ")
	}

	return strings.TrimSpace(content)
}

// CreateExcerpt generates a display excerpt (first 15 + "..." + last 15 chars)
func CreateExcerpt(content string) string {
	// Convert to single line first for display
	singleLine := ConvertToSingleLine(content)

	if len(singleLine) <= 33 { // 15 + 3 ("...") + 15
		return singleLine
	}

	first := singleLine[:15]
	last := singleLine[len(singleLine)-15:]
	return first + "..." + last
}
