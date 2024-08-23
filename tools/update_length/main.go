package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/hultan/crypto"
	"github.com/hultan/softimdb/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MovieOMDB struct {
	Runtime string `json:"Runtime"`
}

const configFile = "/home/per/.config/softteam/softimdb/config.json"

func main() {
	cnf, err := config.LoadConfig(configFile)
	if err != nil {
		panic(err)
	}

	db, err := openDatabase(cnf)
	if err != nil {
		panic(err)
	}

	database, err := db.DB()
	if err != nil {
		panic(err)
	}

	movies, err := getMoviesWithoutLength(database)
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(len(movies))

	m := sync.Mutex{}

	for _, movie := range movies {
		go func(key string) {
			info, err := getMovieInfo(key)
			if err != nil {
				fmt.Println(movie, err)
				return
			}

			length, err := getMovieLength(info)
			if err != nil {
				fmt.Println(movie, info, err)
				return
			}

			m.Lock()
			err = updateMovieLength(database, key, length)
			if err != nil {
				fmt.Println("Error updating movie length:", err)
				return
			}
			wg.Done()
			m.Unlock()
		}(movie)
	}

	wg.Wait()
}

func getMovieInfo(movieId string) (string, error) {
	// Define the OMDB API URL with your API key
	apiKey := "xxxx"
	apiUrl := fmt.Sprintf("http://www.omdbapi.com/?apikey=%s&i=%s", apiKey, movieId)

	// Make an HTTP GET request to the API
	response, err := http.Get(apiUrl)
	if err != nil {
		fmt.Println("Error making HTTP request:", err)
		return "", err
	}
	defer response.Body.Close()

	// Read the response body
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return "", err
	}

	return string(responseBody), nil
}

func getMovieLength(info string) (int, error) {
	m := &MovieOMDB{}
	err := json.NewDecoder(strings.NewReader(info)).Decode(m)
	if err != nil {
		return 0, err
	}

	fmt.Println(m.Runtime)

	if m.Runtime != "N/A" {
		i, err := strconv.Atoi(strings.TrimSuffix(m.Runtime, " min"))
		if err != nil {
			return -1, nil
		}
		return i, nil
	} else {
		return -1, nil
	}
}

func getMoviesWithoutLength(db *sql.DB) ([]string, error) {
	// Query the database
	query := `
        SELECT imdb_id
        FROM softimdb.movies
        WHERE length = -1
    `

	rows, err := db.Query(query)
	if err != nil {
		fmt.Println("Error executing query:", err)
		return nil, err
	}
	defer rows.Close()

	if err = rows.Err(); err != nil {
		fmt.Println("Error iterating through rows:", err)
		return nil, err
	}

	// Slice to store the results
	var imdbIDs []string

	// Iterate through the rows and store the imdb_id values in the slice
	for rows.Next() {
		var imdbID string
		if err = rows.Scan(&imdbID); err != nil {
			fmt.Println("Error scanning row:", err)
			return nil, err
		}
		imdbIDs = append(imdbIDs, imdbID)
	}

	return imdbIDs, nil
}

func openDatabase(config *config.Config) (*gorm.DB, error) {
	passwordDecrypted, err := crypto.Decrypt(config.Database.Password)
	if err != nil {
		return nil, err
	}

	var connectionString = fmt.Sprintf(
		"%s:%s@tcp(%s:%v)/%s?parseTime=True",
		config.Database.User,
		passwordDecrypted,
		config.Database.Server,
		config.Database.Port,
		config.Database.Database,
	)

	db, err := gorm.Open(
		mysql.New(
			mysql.Config{
				DriverName: "mysql",
				DSN:        connectionString,
			},
		), &gorm.Config{},
	)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func updateMovieLength(db *sql.DB, key string, length int) error {
	// Prepare the SQL statement for updating the movie's length
	query := `
        UPDATE softimdb.movies
        SET length = ?
        WHERE imdb_id = ?
    `

	// Execute the SQL statement
	_, err := db.Exec(query, length, key)
	if err != nil {
		return err
	}

	return nil
}
