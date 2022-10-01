package imdb2

import (
	"errors"
	"io"
	"os"
	"path"
)

// IMDB API Key
const testApiKey = "k_12345678"
const apiKeyFileName = ".imdb_api_key"

var InvalidKeyError = errors.New("apikey is invalid")
var apiKeyLength = len(testApiKey)

type ApiKeyManager struct {
	apiKey string
}

func NewApiKeyManager(key string) (*ApiKeyManager, error) {
	if !validateKey(key) {
		return nil, InvalidKeyError
	}
	return &ApiKeyManager{apiKey: key}, nil
}

func NewApiKeyManagerFromStandardPath() (*ApiKeyManager, error) {
	userHome, err := getUserHome()
	if err != nil {
		return nil, err
	}
	apiKeyFile := path.Join(userHome, apiKeyFileName)
	return NewApiKeyManagerFromPath(apiKeyFile)
}

func NewApiKeyManagerFromPath(apiKeyFilePath string) (*ApiKeyManager, error) {
	key, err := getApiKeyFromFile(apiKeyFilePath)
	if err != nil {
		return nil, err
	}
	if !validateKey(key) {
		return nil, InvalidKeyError
	}
	return &ApiKeyManager{apiKey: key}, nil
}

func (a *ApiKeyManager) GetApiKey() string {
	return a.apiKey
}

func validateKey(key string) bool {
	if len(key) != apiKeyLength {
		return false
	}
	if key[:2] != "k_" {
		return false
	}
	// TODO : Test api key?
	return true
}

// getApiKeyFromFile : Return the IMDB api key from the .imdb_api_key file in
// the path provided
func getApiKeyFromFile(dir string) (string, error) {
	// Open path
	file, err := os.Open(dir)
	if err != nil {
		return "", err
	}

	// Read api key from api key file
	text, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	// Return the first 10 characters
	if len(text) > apiKeyLength {
		text = text[:apiKeyLength]
	}

	return string(text), nil
}

// getUserHome : Returns the current users home directory
func getUserHome() (string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return userHome, nil
}
