package imdb

import (
	"testing"

	"github.com/hultan/softimdb/internal/data"
	"github.com/stretchr/testify/assert"
)

func TestManager_GetMovie(t *testing.T) {
	url := "https://www.imdb.com/title/tt0425151/?ref_=nv_sr_srsg_3_tt_8_nm_0_in_0_q_Jimmy%2520and%2520"
	manager := ManagerNew()
	movie, err := manager.GetMovie(url)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, movie.Title, "Jimmy and Judy")
	assert.Equal(t, movie.Runtime, 99)
	assert.Equal(t, movie.Year, 2006)
	assert.Equal(t, movie.Rating, "5.8")
	assert.Equal(t, movie.StoryLine, "In the Kentucky suburbs of Cincinnati, "+
		"social misfit Jimmy Wright always has his video camera - at his psychiatrist's, spying on his parents in their bedroom, and watching high-school senior, Judy Oaks-Kellen. He rescues Judy from a teacher and students who tease and torment her, and showing her his video tape of revenge kick-starts their friendship, which is soon in an overdrive of romance, sex, and pleasure. Jimmy is in and out of mental institutions, and before long, he and Judy are on the run. Cocaine, guns, and a commune of other misfits figure in their flight. How far can their love take them? It's all on video. —<jhailey@hotmail.com>")
	assert.Equal(t, len(movie.Genres), 3)
	assert.Equal(t, movie.Genres[0], "Crime")
	assert.Equal(t, movie.Genres[1], "Drama")
	assert.Equal(t, movie.Genres[2], "Thriller")
	assert.Equal(t, len(movie.Persons), 22)
	assert.Equal(t, movie.Persons[0].Name, "Randall Rubin")
	assert.Equal(t, movie.Persons[0].Type, data.Director)
	assert.Equal(t, movie.Persons[1].Name, "Jonathan Schroder")
	assert.Equal(t, movie.Persons[1].Type, data.Director)
	assert.Equal(t, movie.Persons[2].Name, "Randall Rubin")
	assert.Equal(t, movie.Persons[2].Type, data.Writer)
	assert.Equal(t, movie.Persons[3].Name, "Jonathan Schroder")
	assert.Equal(t, movie.Persons[3].Type, data.Writer)
	assert.Equal(t, movie.Persons[4].Name, "Edward Furlong")
	assert.Equal(t, movie.Persons[4].Type, data.Actor)
}

func TestCalcRuntime(t *testing.T) {
	manager := &Manager{}

	tests := []struct {
		input       string
		expected    int
		expectError bool
	}{
		{"1h 30m", 90, false},
		{"2h", 120, false},
		{"95m", 95, false},
		{" 1h 0m ", 60, false},
		{"1h 75m", 135, false},
		{"0h 0m", -1, true},
		{"", -1, true},
		{"abc", -1, true},
		{"1hr 30min", -1, true}, // unsupported format
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := manager.calcRuntime(test.input)
			if test.expectError {
				if err == nil {
					t.Errorf("expected error for input %q, got none", test.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %q: %v", test.input, err)
				}
				if result != test.expected {
					t.Errorf("input %q: expected %d, got %d", test.input, test.expected, result)
				}
			}
		})
	}
}
