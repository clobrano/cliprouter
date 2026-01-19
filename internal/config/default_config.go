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

# Available environment variable:
#   ${CLIP} - Contains the clipboard content (expanded by the shell)
#
# Configuration:
#   - name: Display name shown in the TUI (required)
#   - command: Command to execute (required, can be multiline for scripts)
#   - timeout: Timeout in seconds (optional, defaults to --timeout flag or 3s)
#   - notify_before: "message" - Notification before command starts (optional)
#   - notify_after: "message" - Notification after command completes (optional)
#   - notify_on_error: "message" - Notification when command fails (optional)
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
#   notify_before: "Starting data processing..."
#   notify_after: "Processing completed. Output: ${OUTPUT_FILE}"
#   notify_on_error: "Processing failed: ${ERROR}"
#   command: |
#     #!/bin/bash
#     # Compute dynamic output file
#     OUTPUT_FILE="/tmp/output-$(date +%s).txt"
#     echo "Processing: ${CLIP}" > "$OUTPUT_FILE"
#     # Expose variable to notification
#     echo "OUTPUT_FILE=$OUTPUT_FILE" >> "$CLIPROUTER_ENV_FILE"
#
# Notifications:
#   - Notification title is always the script name
#   - Message supports variable substitution
#   - Three notification types:
#     notify_before: "message" - Sent before command starts
#     notify_after: "message" - Sent when command succeeds (exit code 0)
#     notify_on_error: "message" - Sent when command fails (exit code != 0)
#   - Available variables in notification messages:
#     ${SCRIPT_NAME} - name of the script
#     ${COMMAND} - the command being executed
#     ${EXIT_CODE} - exit code (in notify_after/notify_on_error)
#     ${STDOUT} - command output (in notify_after/notify_on_error)
#     ${STDERR} - error output (in notify_after/notify_on_error)
#     ${ERROR} - error message (in notify_after/notify_on_error)
#     ${ANY_ENV_VAR} - parent process env vars or CLIPROUTER_ENV_FILE vars
#   - Parent process environment variables are accessible (e.g., ${USER}, ${HOME})
#   - Dynamic vars: Write to $CLIPROUTER_ENV_FILE in your command:
#     echo "VAR_NAME=value" >> "$CLIPROUTER_ENV_FILE"
#   - NOTE: Env vars exported INSIDE commands are not accessible (use CLIPROUTER_ENV_FILE)
`
}
