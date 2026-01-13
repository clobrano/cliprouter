package config

// DefaultConfigContent returns the default YAML configuration content
func DefaultConfigContent() string {
	return `# Clipboard Router Configuration

scripts:
  # Example 1: Simple command with clipboard placeholder
  - name: Echo Clipboard
    command: echo ${CLIP}

  # Example 2: Save to a file with timestamp
  - name: Save to Notes
    command: |
      echo "$(date): ${CLIP}" >> ~/cliprouter-notes.txt

# Available placeholder:
#   ${CLIP} - Will be replaced with the clipboard content
#
# Configuration:
#   - name: Display name shown in the TUI (required)
#   - command: Command to execute (required, can be multiline for scripts)
#   - timeout: Timeout in seconds (optional, defaults to --timeout flag or 3s)
#   - notify_before: Desktop notification before execution (optional)
#   - notify_after: Desktop notification after successful execution (optional)
#   - notify_on_error: Desktop notification on error (optional)
#
# Examples:
#
# - name: Read It Later
#   command: readitlater ${CLIP}
#
# - name: Bookmark URL
#   command: bookmark --url ${CLIP}
#
# - name: Long Running Process
#   timeout: 300  # 5 minutes
#   command: |
#     #!/bin/bash
#     echo "Processing: ${CLIP}"
#     # Your script here
#
# - name: AI Text Editor
#   command: ai-editor.sh ${CLIP}
#   notify_before:
#     message: "Starting AI editing..."
#   notify_after:
#     message: "Editing completed!"
#   notify_on_error:
#     message: "Editing failed: ${ERROR}"
`
}
