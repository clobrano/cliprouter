package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/clobrano/cliprouter/internal/clipboard"
	"github.com/clobrano/cliprouter/internal/config"
	"github.com/clobrano/cliprouter/internal/executor"
	"github.com/clobrano/cliprouter/internal/logger"
	"github.com/clobrano/cliprouter/internal/ui"
)

var (
	configPath = flag.String("config", "", "Path to configuration file")
	verbose    = flag.Bool("v", false, "Enable verbose logging")
	dryRun     = flag.Bool("dry-run", false, "Show command without executing")
	timeoutSec = flag.Int("timeout", 3, "Script execution timeout in seconds")
)

func main() {
	flag.Parse()

	// Initialize logger
	if err := logger.InitLogger(*verbose); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to initialize logger: %v\n", err)
		os.Exit(2)
	}
	defer logger.Close()

	logger.LogInfo("Clipboard Router starting")

	// Get config path
	cfgPath, err := config.GetConfigPath(*configPath)
	if err != nil {
		logger.LogError("Failed to get config path: %v", err)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}

	// Ensure config exists (create default if not)
	if err := config.EnsureConfigExists(cfgPath); err != nil {
		logger.LogError("Failed to ensure config exists: %v", err)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}

	// Load configuration
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		logger.LogError("Failed to load config: %v", err)
		fmt.Fprintf(os.Stderr, "Error: Failed to load configuration: %v\n", err)
		fmt.Fprintf(os.Stderr, "Config file: %s\n", cfgPath)
		os.Exit(1)
	}
	logger.LogInfo("Loaded %d scripts from config", len(cfg.Scripts))

	// Read clipboard
	clipContent, err := clipboard.GetClipboard()
	if err != nil {
		logger.LogError("Failed to read clipboard: %v", err)
		fmt.Fprintf(os.Stderr, "Error: Failed to read clipboard: %v\n", err)
		fmt.Fprintf(os.Stderr, "\nTip: On Linux, ensure you have one of the following installed:\n")
		fmt.Fprintf(os.Stderr, "  - xclip (X11)\n")
		fmt.Fprintf(os.Stderr, "  - xsel (X11)\n")
		fmt.Fprintf(os.Stderr, "  - wl-clipboard (Wayland)\n")
		os.Exit(2)
	}

	// Check if clipboard is empty
	if clipboard.IsEmpty(clipContent) {
		logger.LogError("Clipboard is empty")
		fmt.Fprintf(os.Stderr, "Error: Clipboard is empty. Copy some text and try again.\n")
		os.Exit(1)
	}

	// Convert to single line for processing
	singleLineContent := clipboard.ConvertToSingleLine(clipContent)
	logger.LogDebug("Clipboard content length: %d characters", len(singleLineContent))

	// Create excerpt for display
	excerpt := clipboard.CreateExcerpt(clipContent)

	// Create and run TUI
	model := ui.NewModel(cfg.Scripts, excerpt)
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		logger.LogError("TUI error: %v", err)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(2)
	}

	// Get final model state
	m := finalModel.(ui.Model)

	// Check if user quit without selecting
	if m.IsQuitting() {
		logger.LogInfo("User quit without selecting a script")
		os.Exit(0)
	}

	// Get selected script
	selectedScript := m.GetSelectedScript()
	if selectedScript == nil {
		logger.LogInfo("No script selected")
		os.Exit(0)
	}

	logger.LogInfo("User selected script: %s", selectedScript.Name)

	// Handle dry-run mode
	if *dryRun {
		// Show what would be executed
		command := executor.SubstituteClipboard(selectedScript.Command, singleLineContent)
		fmt.Printf("Would execute script: %s\n", selectedScript.Name)
		fmt.Printf("Command: %s\n", command)
		os.Exit(0)
	}

	// Execute the selected script and wait for it to complete
	// The executor will block until the process exits or times out
	// Use script-specific timeout if set, otherwise use global timeout
	timeout := time.Duration(*timeoutSec) * time.Second
	if selectedScript.Timeout > 0 {
		timeout = time.Duration(selectedScript.Timeout) * time.Second
		logger.LogInfo("Using script-specific timeout: %d seconds", selectedScript.Timeout)
	}
	stdout, stderr, err := executor.Execute(*selectedScript, singleLineContent, timeout)

	// Log output
	if stdout != "" {
		logger.LogInfo("Script stdout: %s", stdout)
	}
	if stderr != "" {
		logger.LogInfo("Script stderr: %s", stderr)
	}

	if err != nil {
		logger.LogError("Script execution failed: %v", err)
		fmt.Fprintf(os.Stderr, "Error: Script execution failed: %v\n", err)
		if stderr != "" {
			fmt.Fprintf(os.Stderr, "stderr: %s\n", stderr)
		}
		os.Exit(1)
	}

	logger.LogInfo("Script executed successfully")
	os.Exit(0)
}
