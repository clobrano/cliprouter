package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"
)

const (
	ConfigFileName = "cliprouter/config.yaml"
)

// NotifyAction represents an interactive notification that requires user response
type NotifyAction struct {
	Prompt    string        `yaml:"prompt"`     // Message to show in the dialog
	OnConfirm *ActionConfig `yaml:"on_confirm"` // Action to execute when user confirms
}

// ActionConfig represents the action to take when user confirms
type ActionConfig struct {
	OpenFile string `yaml:"open_file,omitempty"` // Path to file to open (supports variable substitution)
	Execute  string `yaml:"execute,omitempty"`   // Command to execute (supports variable substitution)
}

// Script represents a single script configuration
type Script struct {
	Name          string        `yaml:"name"`
	Command       string        `yaml:"command"`
	Timeout       int           `yaml:"timeout,omitempty"`          // Timeout in seconds (optional, 0 means use default)
	NotifyBefore  string        `yaml:"notify_before,omitempty"`    // Notification message before command starts
	NotifyAfter   string        `yaml:"notify_after,omitempty"`     // Notification message after command completes successfully
	NotifyOnError string        `yaml:"notify_on_error,omitempty"`  // Notification message when command fails
	NotifyAction  *NotifyAction `yaml:"notify_action,omitempty"`    // Interactive notification after command completes
}

// Config represents the application configuration
type Config struct {
	Scripts []Script `yaml:"scripts"`
}

// LoadConfig reads and parses the YAML configuration file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// validateConfig checks the configuration for errors
func validateConfig(config *Config) error {
	if len(config.Scripts) == 0 {
		return fmt.Errorf("configuration must contain at least one script")
	}

	for i, script := range config.Scripts {
		if strings.TrimSpace(script.Name) == "" {
			return fmt.Errorf("script at index %d is missing a name", i)
		}
		if strings.TrimSpace(script.Command) == "" {
			return fmt.Errorf("script '%s' is missing a command", script.Name)
		}
	}

	return nil
}

// GetConfigPath returns the configuration file path
// If customPath is provided, it validates and returns it
// Otherwise, returns the XDG config directory path
func GetConfigPath(customPath string) (string, error) {
	if customPath != "" {
		// Validate custom path doesn't escape to unsafe locations
		absPath, err := filepath.Abs(customPath)
		if err != nil {
			return "", fmt.Errorf("invalid config path: %w", err)
		}
		return absPath, nil
	}

	// Use XDG config directory
	configPath := filepath.Join(xdg.ConfigHome, ConfigFileName)
	return configPath, nil
}

// CreateDefaultConfig creates a default configuration file at the specified path
func CreateDefaultConfig(path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write default config
	if err := os.WriteFile(path, []byte(DefaultConfigContent()), 0644); err != nil {
		return fmt.Errorf("failed to write default config: %w", err)
	}

	return nil
}

// EnsureConfigExists checks if config exists, creates default if not
func EnsureConfigExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return CreateDefaultConfig(path)
	}
	return nil
}
