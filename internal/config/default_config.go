package config

// DefaultConfigContent returns the default YAML configuration content
func DefaultConfigContent() string {
	return `# Clipboard Router Configuration
# Maximum 10 scripts allowed

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
#   - name: Display name shown in the TUI
#   - command: Command to execute (can be multiline for scripts)
#
# Examples:
#
# - name: Read It Later
#   command: readitlater ${CLIP}
#
# - name: Bookmark URL
#   command: bookmark --url ${CLIP}
#
# - name: Process with Script
#   command: |
#     #!/bin/bash
#     echo "Processing: ${CLIP}"
#     # Your script here
`
}
