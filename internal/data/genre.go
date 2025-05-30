package data

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// Genre represents a movie genre.
type Genre struct {
	Id        int     `gorm:"column:id;primary_key"`
	Name      string  `gorm:"column:name;size:255"`
	IsPrivate bool    `gorm:"column:is_private;"`
	Movies    []Movie `gorm:"-"`
}

var genreCache *GenreCache

// TableName returns the genre table name.
func (t *Genre) TableName() string {
	return "genre"
}

// GetGenres returns all genres
func (d *Database) GetGenres() ([]Genre, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}

	// Get genre id:s for movie
	var genres []Genre
	if result := db.Find(&genres); result.Error != nil {
		return nil, result.Error
	}

	// Fill genre cache
	for i := range genres {
		genreCache.add(&genres[i])
	}

	return genres, nil
}

// getGenreByName returns a genre by name.
func (d *Database) getGenreByName(name string) (*Genre, error) {
	name = strings.Trim(name, " \t\n")

	// Check genre cache
	t := genreCache.getByName(name)
	if t != nil {
		return t, nil
	}

	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}
	genre := Genre{}
	if result := db.Where("name=?", name).First(&genre); result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
		return nil, nil
	}

	return &genre, nil
}

// getOrInsertGenre either returns an existing genre or inserts a new genre and returns it.
func (d *Database) getOrInsertGenre(genre *Genre) (*Genre, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}

	// Check if genre exists
	existingGenre, err := d.getGenreByName(genre.Name)
	if err != nil {
		return nil, err
	}

	// If it does, return it
	if existingGenre != nil {
		return existingGenre, nil
	}

	genre.Name = strings.Trim(genre.Name, " \t\n")

	// If it does not, create it
	if result := db.Create(genre); result.Error != nil {
		return nil, result.Error
	}
	return genre, nil
}

// getGenresForMovie returns a list of genres connected to the given movie.
func (d *Database) getGenresForMovie(movie *Movie) ([]Genre, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}

	// Get genre id:s for movie
	var movieGenres []MovieGenre
	if result := db.Where("movie_id=?", movie.Id).Find(&movieGenres); result.Error != nil {
		return nil, result.Error
	}

	var genres []Genre

	// Get genres for movieGenres
outerLoop:
	for i := range movieGenres {
		// Check genre cache first
		t := genreCache.getById(movieGenres[i].GenreId)
		if t != nil {
			genres = append(genres, *t)
			continue outerLoop
		}

		// Genre did not exist in the genre cache, load it
		// and add it to genre cache
		var genre Genre
		if result := db.Where("id=?", movieGenres[i].GenreId).Find(&genre); result.Error != nil {
			return nil, result.Error
		}
		genreCache.add(&genre)
		genres = append(genres, genre)
	}

	return genres, nil
}

// deleteGenresForMovie deletes all genres for the given movie.
func (d *Database) deleteGenresForMovie(movie *Movie) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}

	// Get genre id:s for movie
	var movieGenres []MovieGenre
	if result := db.Where("movie_id=?", movie.Id).Find(&movieGenres); result.Error != nil {
		return result.Error
	}

	db.Delete(&movieGenres)

	return nil
}
