package config

import (
	"encoding/json"
	"log"
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

	// Open Loader file
	configFile, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = configFile.Close()
		if err != nil {
			log.Fatal(err)
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
