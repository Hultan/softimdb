package nas

import (
	"reflect"
	"testing"
)

func TestRemoveMoviePaths(t *testing.T) {
	manager := ManagerNew(nil)

	tests := []struct {
		name       string
		dirs       []string
		moviePaths []string
		expected   []string
	}{
		{
			name:       "No overlap",
			dirs:       []string{"MovieA", "MovieB"},
			moviePaths: []string{"MovieC"},
			expected:   []string{"MovieA", "MovieB"},
		},
		{
			name:       "Partial overlap",
			dirs:       []string{"MovieA", "MovieB", "MovieC"},
			moviePaths: []string{"MovieB"},
			expected:   []string{"MovieA", "MovieC"},
		},
		{
			name:       "All overlap",
			dirs:       []string{"MovieA", "MovieB"},
			moviePaths: []string{"MovieA", "MovieB"},
			expected:   []string{},
		},
		{
			name:       "Handles empty dirs and empty strings",
			dirs:       []string{"MovieA", "", "MovieB"},
			moviePaths: []string{"MovieB"},
			expected:   []string{"MovieA"},
		},
		{
			name:       "Empty input",
			dirs:       []string{},
			moviePaths: []string{},
			expected:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := manager.removeExistingPaths(tt.dirs, tt.moviePaths)
			if actual == nil {
				actual = []string{}
			}
			if tt.expected == nil {
				tt.expected = []string{}
			}
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, actual)
			}
		})
	}
}
