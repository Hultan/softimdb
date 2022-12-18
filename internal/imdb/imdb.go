package imdb

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type Imdb struct {
	api *ApiKeyManager
}

func NewImdb(manager *ApiKeyManager) *Imdb {
	return &Imdb{manager}
}

// URLs
const searchMoviesURL = "https://imdb-api.com/en/API/SearchMovie/{APIKEY}/{SEARCH}"
const titleURL = "https://imdb-api.com/en/API/Title/{APIKEY}/{ID}/Wikipedia"

func (i *Imdb) SearchMovies(searchString string) (*MovieResults, error) {
	parameters := []parameter{{"{APIKEY}", i.api.GetApiKey()}, {"{SEARCH}", searchString}}
	url := replaceParameters(searchMoviesURL, parameters)

	result, err := makeApiCall(url)
	if err != nil {
		return nil, err
	}

	// Convert to JSON
	movieResults := &MovieResults{}
	err = json.Unmarshal(result, movieResults)
	if err != nil {
		return nil, err
	}

	// Check error message
	if movieResults.ErrorMessage != "" {
		return nil, errors.New(movieResults.ErrorMessage)
	}

	return movieResults, nil
}

func (i *Imdb) Title(id string) (*Movie, error) {
	parameters := []parameter{{"{APIKEY}", i.api.GetApiKey()}, {"{ID}", id}}
	url := replaceParameters(titleURL, parameters)

	result, err := makeApiCall(url)
	if err != nil {
		return nil, err
	}

	// Convert to JSON
	movie := &Movie{}
	err = json.Unmarshal(result, movie)
	if err != nil {
		return nil, err
	}

	// Check error message
	if movie.ErrorMessage != "" {
		return nil, errors.New(movie.ErrorMessage)
	}

	return movie, nil
}

func makeApiCall(url string) ([]byte, error) {
	// Create HTTP client
	client := &http.Client{}

	// Crete GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Perform GET request
	res, err := client.Do(req)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	// Read the entire body (JSON)
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
