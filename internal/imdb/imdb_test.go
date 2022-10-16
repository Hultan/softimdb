package imdb

import (
	"strings"
	"testing"
)

func TestImdb_SearchMovies(t *testing.T) {
	type args struct {
		searchString string
	}

	tests := []struct {
		name string
		args args
	}{
		{"Test 1", args{"gladiator"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := NewApiKeyManagerFromStandardPath()
			if err != nil {
				t.Errorf("NewApiKeyManagerFromStandardPath failed with  err = %v", err)
			}
			i := NewImdb(a)
			got, err := i.SearchMovies(tt.args.searchString)
			if err != nil {
				t.Errorf("SearchMovies failed with err = %v", err)
			}
			if got == nil {
				t.Errorf("SearchMovies returnes nil")
			}
			if got.ErrorMessage != "" {
				t.Errorf("SearchMovies returned error message = %v", err)
			}
			if got.SearchType != "Movie" {
				t.Errorf("SearchMovies returned searchType = %v, want = %v", got.SearchType, "Movie")
			}
			if got.Expression != "gladiator" {
				t.Errorf("SearchMovies returned expression = %v, want = %v", got.Expression, "gladiator")
			}
			if len(got.Results) != 7 {
				t.Errorf("SearchMovies returned wrong number of results = %v, want = %v", len(got.Results), 7)
			}
		})
	}
}

func TestImdb_Title(t *testing.T) {
	type args struct {
		id string
	}

	tests := []struct {
		name string
		args args
	}{
		{"Test Inception", args{"tt1375666"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := NewApiKeyManagerFromStandardPath()
			if err != nil {
				t.Errorf("NewApiKeyManagerFromStandardPath failed with  err = %v", err)
			}
			i := NewImdb(a)
			got, err := i.Title(tt.args.id)
			if err != nil {
				t.Errorf("Title failed with err = %v", err)
			}
			if got == nil {
				t.Errorf("Title returnes nil")
			}
			if got.ErrorMessage != "" {
				t.Errorf("Title returned error message = %v", err)
			}
			if got.Title != "Inception" {
				t.Errorf("Title returned wrong title = %v, want = %v", got.Title, "Inception")
			}
			if got.Type != "Movie" {
				t.Errorf("Title returned wrong type = %v, want = %v", got.Type, "Movie")
			}
			if got.Year != "2010" {
				t.Errorf("Title returned wrong year = %v, want = %v", got.Year, "2010")
			}
			if !strings.Contains(got.ImageURL, "MV5BMjAxMzY3NjcxNF5BMl5BanBnXkFtZTcwNTI5OTM0Mw@@._V1_Ratio0.6762_AL_.jpg") {
				t.Errorf("Title returned wrong image url = %v, want = %v", got.ImageURL, "https://m.media-amazon.com/images/M/MV5BMjAxMzY3NjcxNF5BMl5BanBnXkFtZTcwNTI5OTM0Mw@@._V1_Ratio0.6762_AL_.jpg")
			}
			if !strings.HasPrefix(got.StoryLine, "A thief who steals corporate secrets through") {
				t.Errorf("Title returned wrong story line = %v, want = %v", got.StoryLine, "A thief who steals corporate secrets through")
			}
			if got.Genres != "Action, Adventure, Sci-Fi" {
				t.Errorf("Title returned wrong genres = %v, want = %v", got.Genres, "Action, Adventure, Sci-Fi")
			}
			if got.Rating != "8.8" {
				t.Errorf("Title returned wrong rating = %v, want = %v", got.Rating, "8.8")
			}
		})
	}
}
