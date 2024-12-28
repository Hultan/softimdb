package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/hultan/softimdb/internal/config"
	"github.com/hultan/softimdb/internal/data"
	"github.com/hultan/softimdb/internal/imdb"
)

func main() {
	quitChannel := make(chan bool)

	for {
		go updateMovie(quitChannel)

		// Use a select statement to check the quitChannel
		select {
		case quit := <-quitChannel:
			if !quit {
				// Exit the application if the channel receives false
				println("Quitting application...")
				return
			}
		default:
			// No message received on quitChannel; continue the loop
		}

		r := rand.Intn(60) + 30
		time.Sleep(time.Duration(r) * time.Second)
	}
}

func updateMovie(quitChannel chan bool) {
	db, err := openDatabase()
	if err != nil {
		panic(err)
	}
	defer db.CloseDatabase()

	movies, err := db.SearchMoviesEx("all", "", -1, "id", true)
	if err != nil {
		panic(err)
	}

	if len(movies) == 0 {
		quitChannel <- false
		return
	}

	fmt.Println("=======================")
	fmt.Println(movies[0].Title)
	fmt.Println("=======================")

	manager := imdb.ManagerNew()
	movie, err := manager.GetMovie(movies[0].ImdbUrl)
	if err != nil {
		//panic(err)
	}
	if movie == nil {
		panic(err)
	}

	movies[0].Persons = movie.Persons

	err = db.UpdateMoviePersons(movies[0])
	if err != nil {
		panic(err)
	}

	err = db.SetProcessed(movies[0])
	if err != nil {
		panic(err)
	}

	quitChannel <- true
}

func openDatabase() (*data.Database, error) {
	const configFile = "/home/per/.config/softteam/softimdb/config.json"
	cnf, err := config.LoadConfig(configFile)
	if err != nil {
		return nil, err
	}
	return data.DatabaseNew(false, cnf), nil
}
