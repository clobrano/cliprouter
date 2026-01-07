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
			name: "empty scripts",
			config: &Config{
				Scripts: []Script{},
			},
			expectErr: true,
		},
		{
			name: "too many scripts",
			config: &Config{
				Scripts: make([]Script, 11),
			},
			expectErr: true,
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
		{
			name: "max scripts allowed",
			config: &Config{
				Scripts: make([]Script, 10),
			},
			expectErr: true, // Will fail because scripts don't have names/commands
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For the "max scripts allowed" test, populate with valid data
			if tt.name == "max scripts allowed" {
				for i := 0; i < 10; i++ {
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
