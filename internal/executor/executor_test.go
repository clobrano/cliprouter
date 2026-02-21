package executor

import (
	"testing"
	"time"

	"github.com/clobrano/cliprouter/internal/config"
)

func TestSubstituteClipboard(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		clipboard string
		expected  string
	}{
		{
			name:      "simple substitution",
			command:   "echo ${CLIP}",
			clipboard: "hello",
			expected:  "echo hello",
		},
		{
			name:      "multiple substitutions",
			command:   "echo ${CLIP} and ${CLIP}",
			clipboard: "test",
			expected:  "echo test and test",
		},
		{
			name:      "no substitution",
			command:   "echo hello",
			clipboard: "world",
			expected:  "echo hello",
		},
		{
			name:      "special characters",
			command:   "echo ${CLIP}",
			clipboard: "hello'world",
			expected:  "echo hello'world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := substituteClipboard(tt.command, tt.clipboard)
			if result != tt.expected {
				t.Errorf("substituteClipboard() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestIsMultilineCommand(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{"single line", "echo hello", false},
		{"multiline", "echo hello\necho world", true},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isMultilineCommand(tt.command)
			if result != tt.expected {
				t.Errorf("isMultilineCommand(%q) = %v, expected %v", tt.command, result, tt.expected)
			}
		})
	}
}

func TestExecute(t *testing.T) {
	tests := []struct {
		name        string
		script      config.Script
		clipboard   string
		expectError bool
	}{
		{
			name: "simple echo",
			script: config.Script{
				Name:    "Echo Test",
				Command: "echo ${CLIP}",
			},
			clipboard:   "hello",
			expectError: false,
		},
		{
			name: "invalid command",
			script: config.Script{
				Name:    "Invalid",
				Command: "nonexistentcommand12345",
			},
			clipboard:   "test",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeout := 5 * time.Second
			_, _, err := Execute(tt.script, tt.clipboard, timeout)
			if (err != nil) != tt.expectError {
				t.Errorf("Execute() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestHasShebang(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{"bash shebang", "#!/bin/bash\necho test", true},
		{"env bash shebang", "#!/usr/bin/env bash\necho test", true},
		{"with whitespace", "  #!/bin/sh\necho test", true},
		{"no shebang", "echo test", false},
		{"multiline no shebang", "echo test\necho more", false},
		{"comment not shebang", "# comment\necho test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasShebang(tt.command)
			if result != tt.expected {
				t.Errorf("hasShebang(%q) = %v, expected %v", tt.command, result, tt.expected)
			}
		})
	}
}
