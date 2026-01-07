# Clipboard Router

A fast, keyboard-driven CLI tool with a Terminal User Interface (TUI) that routes clipboard content to pre-configured scripts. Perfect for quickly processing URLs, text snippets, or any clipboard content with custom commands.

## Features

- **Fast TUI**: Navigate scripts with arrow keys or fuzzy search
- **Flexible Configuration**: YAML-based script configuration with multiline support
- **Cross-Platform**: Works on Linux (X11/Wayland) and macOS
- **Safe Execution**: Proper shell escaping and timeout protection
- **Dry-Run Mode**: Preview commands before execution
- **XDG Compliant**: Follows XDG Base Directory specification

## Installation

### Prerequisites

**Go 1.24.0 or later**

**Platform-specific clipboard utilities:**

- **Linux (X11)**: Install `xclip` or `xsel`
  ```bash
  # Debian/Ubuntu
  sudo apt install xclip

  # Fedora
  sudo dnf install xclip

  # Arch
  sudo pacman -S xclip
  ```

- **Linux (Wayland)**: Install `wl-clipboard`
  ```bash
  # Debian/Ubuntu
  sudo apt install wl-clipboard

  # Fedora
  sudo dnf install wl-clipboard

  # Arch
  sudo pacman -S wl-clipboard
  ```

- **macOS**: No additional dependencies (uses built-in `pbpaste`)

### Build from Source

```bash
git clone <repository-url>
cd cliprouter
go build -o cliprouter
sudo mv cliprouter /usr/local/bin/
```

## Quick Start

1. **Run the application** (it will create a default config on first run):
   ```bash
   cliprouter
   ```

2. **Edit the configuration file**:
   ```bash
   # Linux
   nano ~/.config/cliprouter/config.yaml

   # macOS
   nano ~/Library/Application\ Support/cliprouter/config.yaml
   ```

3. **Copy some text** to your clipboard

4. **Run cliprouter** again and select your script!

## Configuration

The configuration file is a YAML file located at:
- **Linux**: `~/.config/cliprouter/config.yaml` (or `$XDG_CONFIG_HOME/cliprouter/config.yaml`)
- **macOS**: `~/Library/Application Support/cliprouter/config.yaml`

### Configuration Format

```yaml
scripts:
  - name: Script Name
    command: command to execute ${CLIP}
```

- **Maximum 10 scripts** allowed
- **`${CLIP}`** placeholder is replaced with clipboard content
- **Multiline commands** are supported

### Example Configuration

```yaml
scripts:
  # Simple command
  - name: Read It Later
    command: readitlater ${CLIP}

  # Command with arguments
  - name: Bookmark URL
    command: bookmark --path ~/bookmarks.yaml --input ${CLIP}

  # Multiline script
  - name: Process URL
    command: |
      #!/bin/bash
      echo "Processing: ${CLIP}"
      curl -X POST https://api.example.com/process \
        -d "url=${CLIP}"

  # Save to file with timestamp
  - name: Save to Notes
    command: |
      echo "$(date): ${CLIP}" >> ~/notes/$(date +%Y-%m-%d).txt
```

## Usage

### Basic Usage

```bash
# Run with default config
cliprouter

# Use custom config file
cliprouter --config /path/to/config.yaml

# Enable verbose logging
cliprouter -v

# Preview command without executing
cliprouter --dry-run

# Custom timeout (default: 30 seconds)
cliprouter --timeout 60
```

### TUI Controls

- **↑/↓ or k/j**: Navigate scripts
- **Type**: Fuzzy search scripts
- **Enter**: Execute selected script
- **Backspace**: Remove search character
- **ESC**: Quit without executing

### Command-Line Flags

- `--config <path>`: Specify custom configuration file
- `-v`, `--verbose`: Enable verbose logging to stderr
- `--dry-run`: Show command without executing
- `--timeout <seconds>`: Script execution timeout (default: 30)
- `-h`, `--help`: Show help message

## Logs

Logs are stored at:
- **Linux**: `~/.local/share/cliprouter/cliprouter.log` (or `$XDG_DATA_HOME/cliprouter/cliprouter.log`)
- **macOS**: `~/Library/Application Support/cliprouter/cliprouter.log`

Use `-v` flag to also output logs to stderr.

## Examples

### Save URLs to Read-It-Later Service

```yaml
scripts:
  - name: Pocket
    command: pocket-cli add ${CLIP}
```

### Format Code Snippets

```yaml
scripts:
  - name: Format JSON
    command: echo ${CLIP} | jq .

  - name: Format Go Code
    command: |
      echo ${CLIP} | gofmt
```

### Quick Notes

```yaml
scripts:
  - name: Daily Note
    command: |
      echo "${CLIP}" >> ~/notes/$(date +%Y-%m-%d).txt

  - name: Inbox
    command: |
      echo "- [ ] ${CLIP}" >> ~/inbox.md
```

### Process URLs

```yaml
scripts:
  - name: Archive URL
    command: waybackmachine ${CLIP}

  - name: Take Screenshot
    command: screenshot-url ${CLIP}
```

## Troubleshooting

### Clipboard Error on Linux

**Error**: "Failed to read clipboard"

**Solution**: Install the appropriate clipboard utility:
- X11: `xclip` or `xsel`
- Wayland: `wl-clipboard`

### Empty Clipboard Error

**Error**: "Clipboard is empty"

**Solution**: Copy some text before running cliprouter.

### Configuration Errors

**Error**: "Failed to parse YAML"

**Solution**:
- Check YAML syntax (proper indentation, no tabs)
- Ensure each script has both `name` and `command` fields
- Maximum 10 scripts allowed

### Script Timeout

**Error**: "Script execution timed out"

**Solution**: Increase timeout with `--timeout` flag:
```bash
cliprouter --timeout 60
```

### Script Execution Failed

**Check logs** for details:
```bash
# Linux
tail -f ~/.local/share/cliprouter/cliprouter.log

# macOS
tail -f ~/Library/Application\ Support/cliprouter/cliprouter.log
```

**Run with verbose mode**:
```bash
cliprouter -v
```

**Test command with dry-run**:
```bash
cliprouter --dry-run
```

## Development

### Run Tests

```bash
# Run all tests
go test ./...

# Run tests for specific package
go test -v ./internal/config

# Run with coverage
go test -cover ./...
```

### Code Quality

```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Run linter (if golangci-lint is installed)
golangci-lint run
```

## Security Considerations

- Clipboard content is **properly escaped** before execution to prevent command injection
- Scripts execute with **timeout protection** (default 30 seconds)
- Configuration limited to **maximum 10 scripts** to prevent resource exhaustion
- Always **review scripts** in your configuration before executing

## License

[Add your license here]

## Contributing

[Add contribution guidelines here]
