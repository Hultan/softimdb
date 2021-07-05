package main

import (
	"bufio"
	"fmt"
	"github.com/hultan/softimdb/internal/data"
	"github.com/hultan/softimdb/internal/imdb"
	"github.com/hultan/softimdb/internal/softimdb"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const (
	ApplicationId    = "se.softteam.softimdb"
	ApplicationFlags = glib.APPLICATION_FLAGS_NONE
)

func main() {

	//test()
	//return

	// Create a new application
	application, err := gtk.ApplicationNew(ApplicationId, ApplicationFlags)
	softimdb.ErrorCheckWithPanic(err, "Failed to create GTK Application")

	mainForm := softimdb.NewMainWindow()
	// Hook up the activate event handler
	_ = application.Connect("activate", mainForm.OpenMainWindow)

	// Start the application (and exit when it is done)
	os.Exit(application.Run(nil))
}

func test() {
	db := data.DatabaseNew(false)

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Add movie (URL) : ")
		url, _ := reader.ReadString('\n')

		url = strings.Replace(url, "\n", "", -1)

		if strings.Compare("list", strings.ToLower(url)) == 0 {
			PrintAllMovies(db)
			continue
		}

		if strings.Compare("exit", strings.ToLower(url)) == 0 {
			break
		}

		AddMovie(db, url)
	}

	db.CloseDatabase()
}

func ShowMoviePoster(db *data.Database, movieId int) {
	movie, err := db.GetMovie(movieId)
	if err != nil {
		panic(err)
	}

	name := CreateValidFileName(movie.Title)
	fmt.Println(name)

	err = ioutil.WriteFile(name, *movie.Image, 0644)
	if err != nil {
		panic(err)
	}
}

func CreateValidFileName(name string) string {
	var newName string=""

	for i:=range name {
		c:=name[i]
		if (c>='A' && c<='Z') || (c>='a' && c<='z')  || (c>='0' && c<='9') {
			newName += string(c)
		}
	}

	return "/home/per/temp/movieposters/" + newName + ".jpg"
}

func AddMovie(db *data.Database, url string) int {
	movie := data.Movie{ImdbUrl: url}

	imdbManager := imdb.ManagerNew()
	err := imdbManager.GetMovieInfo(&movie)
	if err != nil {
		panic(err)
	}

	err = db.InsertMovie(&movie)
	if err != nil {
		panic(err)
	}

	ShowMoviePoster(db, movie.Id)

	return movie.Id
}

func PrintAllMovies(db *data.Database) {
	movies, _ := db.GetAllMovies()

	for i := range movies {
		movie := movies[i]
		fmt.Println("------------------------------------------")
		fmt.Println("Title : ", movie.Title)
		fmt.Println("------------------------------------------")
		fmt.Println("Year : ", movie.Year)
		fmt.Println("Rating : ", movie.ImdbRating)

		fmt.Print("Tags : ")
		for t := range movie.Tags {
			if t != 0 {
				fmt.Print(", ")
			}
			fmt.Print(movie.Tags[t].Name)
		}
		fmt.Println()
		fmt.Println()
		fmt.Println()
	}
}
