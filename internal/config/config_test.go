package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		expectErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Scripts: []Script{
					{Name: "Test", Command: "echo test"},
				},
			},
			expectErr: false,
		},
		{
			name: "valid config with timeout",
			config: &Config{
				Scripts: []Script{
					{Name: "Test", Command: "echo test", Timeout: 60},
				},
			},
			expectErr: false,
		},
		{
			name: "empty scripts",
			config: &Config{
				Scripts: []Script{},
			},
			expectErr: true,
		},
		{
			name: "many scripts allowed",
			config: &Config{
				Scripts: make([]Script, 50),
			},
			expectErr: true, // Will fail because scripts don't have names/commands
		},
		{
			name: "missing name",
			config: &Config{
				Scripts: []Script{
					{Name: "", Command: "echo test"},
				},
			},
			expectErr: true,
		},
		{
			name: "missing command",
			config: &Config{
				Scripts: []Script{
					{Name: "Test", Command: ""},
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For the "many scripts allowed" test, populate with valid data
			if tt.name == "many scripts allowed" {
				for i := 0; i < 50; i++ {
					tt.config.Scripts[i] = Script{Name: "Test", Command: "echo test"}
				}
				tt.expectErr = false
			}

			err := validateConfig(tt.config)
			if (err != nil) != tt.expectErr {
				t.Errorf("validateConfig() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}

func TestCreateDefaultConfig(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create default config
	err := CreateDefaultConfig(configPath)
	if err != nil {
		t.Fatalf("CreateDefaultConfig() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Verify content is not empty
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	if len(data) == 0 {
		t.Error("Config file is empty")
	}

	// Verify it can be loaded
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Errorf("Failed to load default config: %v", err)
	}

	if len(cfg.Scripts) == 0 {
		t.Error("Default config has no scripts")
	}
}

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	// Write valid config
	validYAML := `scripts:
  - name: Test Script
    command: echo test
  - name: Another Script
    command: |
      echo multi
      echo line
`
	if err := os.WriteFile(configPath, []byte(validYAML), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if len(cfg.Scripts) != 2 {
		t.Errorf("Expected 2 scripts, got %d", len(cfg.Scripts))
	}

	if cfg.Scripts[0].Name != "Test Script" {
		t.Errorf("Expected script name 'Test Script', got %q", cfg.Scripts[0].Name)
	}
}

func TestLoadConfigWithNotifications(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config-notifications.yaml")

	// Write config with notification settings
	yamlWithNotifications := `scripts:
  - name: Test Script
    command: echo test
    notify_before:
      message: "Starting execution"
    notify_after:
      message: "Execution completed"
    notify_on_error:
      message: "Execution failed: ${ERROR}"
`
	if err := os.WriteFile(configPath, []byte(yamlWithNotifications), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if len(cfg.Scripts) != 1 {
		t.Fatalf("Expected 1 script, got %d", len(cfg.Scripts))
	}

	script := cfg.Scripts[0]

	// Verify NotifyBefore
	if script.NotifyBefore == nil {
		t.Error("NotifyBefore should not be nil")
	} else if script.NotifyBefore.Message != "Starting execution" {
		t.Errorf("Expected NotifyBefore.Message 'Starting execution', got %q", script.NotifyBefore.Message)
	}

	// Verify NotifyAfter
	if script.NotifyAfter == nil {
		t.Error("NotifyAfter should not be nil")
	} else if script.NotifyAfter.Message != "Execution completed" {
		t.Errorf("Expected NotifyAfter.Message 'Execution completed', got %q", script.NotifyAfter.Message)
	}

	// Verify NotifyOnError
	if script.NotifyOnError == nil {
		t.Error("NotifyOnError should not be nil")
	} else if script.NotifyOnError.Message != "Execution failed: ${ERROR}" {
		t.Errorf("Expected NotifyOnError.Message 'Execution failed: ${ERROR}', got %q", script.NotifyOnError.Message)
	}
}

func TestLoadConfigWithoutNotifications(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config-no-notifications.yaml")

	// Write config without notification settings
	yamlWithoutNotifications := `scripts:
  - name: Test Script
    command: echo test
`
	if err := os.WriteFile(configPath, []byte(yamlWithoutNotifications), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Load config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if len(cfg.Scripts) != 1 {
		t.Fatalf("Expected 1 script, got %d", len(cfg.Scripts))
	}

	script := cfg.Scripts[0]

	// Verify notification fields are nil when not specified
	if script.NotifyBefore != nil {
		t.Error("NotifyBefore should be nil when not specified")
	}

	if script.NotifyAfter != nil {
		t.Error("NotifyAfter should be nil when not specified")
	}

	if script.NotifyOnError != nil {
		t.Error("NotifyOnError should be nil when not specified")
	}
}
