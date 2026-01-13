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
#   - notify_before: Notification before command starts (optional)
#   - notify_after: Notification after command completes (optional)
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
#   notify_before:
#     title: "Processing Started"
#     message: "Running ${SCRIPT_NAME}"
#   notify_after:
#     title: "Processing Complete"
#     message: "Finished with exit code ${EXIT_CODE}"
#   command: |
#     #!/bin/bash
#     echo "Processing: ${CLIP}"
#     # Your script here
#
# Notification Variables:
#   ${SCRIPT_NAME} - name of the script
#   ${COMMAND} - the command being executed
#   ${EXIT_CODE} - exit code (in notify_after)
#   ${STDOUT} - command output (in notify_after)
#   ${STDERR} - error output (in notify_after)
#   ${ERROR} - error message (in notify_after)
#   ${ANY_ENV_VAR} - any environment variable
`
}
