package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

type Config struct {
	RootDir  string          `json:"rootDir"`
	Database DatabaseSection `json:"database"`
}

type DatabaseSection struct {
	Server   string `json:"server"`
	Database string `json:"database"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

// LoadConfig : Loads the config
func LoadConfig(path string) (*Config, error) {
	p, err := expandPath(path)
	if err != nil {
		return nil, err
	}

	if _, statErr := os.Stat(p); os.IsNotExist(statErr) {
		return nil, fmt.Errorf("config file does not exist: %s", p)
	}

	// Open Loader file
	configFile, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file %s: %w", p, err)
	}
	defer func() {
		if closeErr := configFile.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close failed: %w", closeErr)
		}
	}()

	config := &Config{}
	err = json.NewDecoder(configFile).Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// expandPath expands a path with "~" to the full home directory path
func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		homeDir := usr.HomeDir
		return filepath.Join(homeDir, path[1:]), nil
	}
	return path, nil
}
