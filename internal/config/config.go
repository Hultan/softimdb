package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Nas    string
	Folder string
}

// LoadConfig : Loads the config
func LoadConfig(path string) (*Config, error) {
	// Open Loader file
	configFile, err := os.Open(path)

	// Handle errors
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	defer func() {
		err = configFile.Close()
	}()

	config := &Config{}

	// Parse the JSON document
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
