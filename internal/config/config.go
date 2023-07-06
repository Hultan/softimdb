package config

import (
	"encoding/json"
	"log"
	"os"
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
	// Open Loader file
	configFile, err := os.Open(path)
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
