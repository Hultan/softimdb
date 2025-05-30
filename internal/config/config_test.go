package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("could not get home directory: %v", err)
	}

	// Construct valid paths
	tildePath := "~/.config/softteam/softimdb/config.json"
	fullPath := filepath.Join(home, ".config/softteam/softimdb/config.json")

	tests := []struct {
		name       string
		inputPath  string
		shouldPass bool
	}{
		{"Valid full path", fullPath, true},
		{"Valid tilde path", tildePath, true},
		{"Nonexistent file", "/invalid/path/to/config.json", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := LoadConfig(tt.inputPath)
			if tt.shouldPass {
				if err != nil {
					t.Errorf("expected success, got error: %v", err)
				}
				if cfg == nil {
					t.Errorf("expected config to be loaded, got nil")
				}
			} else {
				if err == nil {
					t.Errorf("expected error for path %q, got nil", tt.inputPath)
				}
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input, expected string
	}{
		{"~/testdir", filepath.Join(home, "testdir")},
		{"/absolute/path", "/absolute/path"},
	}

	for _, test := range tests {
		result, err := expandPath(test.input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != test.expected {
			t.Errorf("expandPath(%q) = %q; expected %q", test.input, result, test.expected)
		}
	}
}
