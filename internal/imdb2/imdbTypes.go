package imdb2

import (
	"os"
	"strconv"
	"strings"
)

type MovieResults struct {
	SearchType   string        `json:"searchType"`
	Expression   string        `json:"expression"`
	Results      []MovieResult `json:"results"`
	ErrorMessage string        `json:"errorMessage"`
}

type MovieResult struct {
	Id          string `json:"id"`
	ResultType  string `json:"resultType"`
	Image       string `json:"image"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type Movie struct {
	Id        string `json:"id"`
	Title     string `json:"title"`
	Type      string `json:"type"`
	Year      string `json:"year"`
	ImageURL  string `json:"image"`
	StoryLine string `json:"plot"`
	Genres    string `json:"genres"`
	Rating    string `json:"imDbRating"`

	ErrorMessage string `json:"errorMessage"`
}

func (m *Movie) GetYear() (int, error) {
	year, err := strconv.Atoi(m.Year)
	if err != nil {
		return -1, err
	}
	return year, nil
}

func (m *Movie) GetRating() (float64, error) {
	rating, err := strconv.ParseFloat(m.Rating, 64)
	if err != nil {
		return 0, err
	}
	return rating, nil
}

func (m *Movie) GetGenres() []string {
	return m.getGenres(m.Genres)
}

func (m *Movie) GetPoster() ([]byte, error) {
	file, err := os.ReadFile(m.ImageURL)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (m *Movie) getGenres(text string) []string {
	var result []string
	genres := strings.Split(text, ",")
	for _, genre := range genres {
		result = append(result, genre)
	}
	return result
}
