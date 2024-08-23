package main

import (
	"log"
	"strings"

	"github.com/hultan/softimdb/internal/config"
	"github.com/hultan/softimdb/internal/data"
)

const configFile = "/home/per/.config/softteam/softimdb/config.json"

func main() {
	// Load config file
	cnf, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}

	// Open database
	database := data.DatabaseNew(false, cnf)

	movies, err := database.SearchMovies("", "", -1, "id asc")
	if err != nil {
		log.Fatal(err)
	}
	for i := range movies {
		movie := movies[i]
		i := strings.Index(movie.ImdbUrl, "/?ref")
		if i >= 0 {
			url := movie.ImdbUrl[:i+1]
			movie.ImdbUrl = url
			err = database.UpdateMovie(movie, false)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	database.CloseDatabase()
}
