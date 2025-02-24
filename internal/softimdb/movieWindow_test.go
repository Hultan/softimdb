package softimdb

import (
	"reflect"
	"testing"
)

func Test_findSimilarMovies(t *testing.T) {
	type args struct {
		newTitle  string
		maxVideos int
	}
	tests := []struct {
		name string
		args args
		want []movie
	}{
		{"Empty title", args{newTitle: ""}, []movie{}},
		{"Gladiator", args{newTitle: "Gladiator II", maxVideos: 2}, []movie{{1000, "Gladiator II"}, {777, "Gladiator"}}},
		{"Inception", args{newTitle: "Inception", maxVideos: 2}, []movie{{1000, "Inception"}, {380, "Interstellar"}}},
		{"The Matrix", args{newTitle: "The Matrix Resurrections", maxVideos: 2}, []movie{{1000, "The Matrix Resurrections"}, {844, "The Matrix Revolutions"}}},
		{"Jurassic Park", args{newTitle: "Jurassic Park", maxVideos: 3}, []movie{{1000, "Jurassic Park"}, {701, "Jurassic World"}, {505, "Jurassic World: Fallen Kingdom"}}},
		{"Jurassic World", args{newTitle: "Jurassic World", maxVideos: 3}, []movie{{1000, "Jurassic World"}, {731, "Jurassic World: Fallen Kingdom"}, {701, "Jurassic Park"}}},
		{"Mad Max: Fury Road", args{newTitle: "Mad Max: Fury Road", maxVideos: 3}, []movie{{1000, "Mad Max: Fury Road"}, {552, "Mad Max Beyond Thunderdome"}, {483, "The Matrix Reloaded"}}},
		{"The Blouse", args{newTitle: "The Blouse", maxVideos: 3}, []movie{{848, "The House"}, {848, "The House"}, {555, "The Godfather"}}},
		{"The Mouse", args{newTitle: "The Mouse", maxVideos: 3}, []movie{{892, "The House"}, {892, "The House"}, {574, "The Godfather"}}},
	}

	titles := getTitles()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findSimilarMovies(tt.args.newTitle, titles, tt.args.maxVideos); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findSimilarMovies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getTitles() []string {
	movies := []string{
		"The House",
		"The House",
		"The Godfather",
		"The Godfather Part II",
		"Inception",
		"Interstellar",
		"The Dark Knight",
		"The Dark Knight Rises",
		"Gladiator",
		"Gladiator II",
		"Avatar",
		"Avatar: The Way of Water",
		"Titanic",
		"Jurassic Park",
		"Jurassic World",
		"Jurassic World: Fallen Kingdom",
		"The Matrix",
		"The Matrix Reloaded",
		"The Matrix Revolutions",
		"The Matrix Resurrections",
		"Mad Max: Fury Road",
		"Mad Max Beyond Thunderdome",
	}

	return movies
}
