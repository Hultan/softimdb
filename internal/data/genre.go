package data

import (
	"errors"
	"fmt"
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

// TableName returns the genre table name.
func (t *Genre) TableName() string {
	return "genre"
}

// GetGenres returns all genres
func (d *Database) GetGenres() ([]Genre, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	// Get genre id:s for movie
	var genres []Genre
	if err := db.Find(&genres).Error; err != nil {
		return nil, fmt.Errorf("failed to query genres: %w", err)
	}

	// Fill genre cache
	for i := range genres {
		d.genreCache.add(&genres[i])
	}

	return genres, nil
}

// getGenreByName returns a genre by name.
func (d *Database) getGenreByName(name string) (*Genre, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil // or return an error if an empty name is invalid
	}

	// Check genre cache first
	if genre := d.genreCache.getByName(name); genre != nil {
		return genre, nil
	}

	db, err := d.getDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	var genre Genre
	result := db.Where("name = ?", name).First(&genre)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Not found is not an error, so return nil, nil
		}
		return nil, fmt.Errorf("database query error: %w", result.Error)
	}

	return &genre, nil
}

// getOrInsertGenre either returns an existing genre or inserts a new genre and returns it.
func (d *Database) getOrInsertGenre(genre *Genre) (*Genre, error) {
	genre.Name = strings.TrimSpace(genre.Name)
	if genre.Name == "" {
		return nil, fmt.Errorf("genre name cannot be empty")
	}

	// Check if the genre already exists
	existingGenre, err := d.getGenreByName(genre.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to query genre: %w", err)
	}
	if existingGenre != nil {
		return existingGenre, nil
	}

	db, err := d.getDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	// Insert new genre
	if result := db.Create(genre); result.Error != nil {
		return nil, fmt.Errorf("failed to create genre: %w", result.Error)
	}

	return genre, nil
}

// getGenresForMovie returns a list of genres connected to the given movie.
func (d *Database) getGenresForMovie(movie *Movie) ([]Genre, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	// Fetch genre IDs associated with the movie
	var movieGenres []MovieGenre
	if result := db.Where("movie_id = ?", movie.Id).Find(&movieGenres); result.Error != nil {
		return nil, fmt.Errorf("failed to query movie genres: %w", result.Error)
	}

	var genres []Genre
	for _, mg := range movieGenres {
		// Attempt to get genre from the cache
		if cachedGenre := d.genreCache.getById(mg.GenreId); cachedGenre != nil {
			genres = append(genres, *cachedGenre)
			continue
		}

		// Not in cache, query the genre from DB
		var genre Genre
		if result := db.First(&genre, mg.GenreId); result.Error != nil {
			return nil, fmt.Errorf("failed to query genre with ID %d: %w", mg.GenreId, result.Error)
		}

		// Add to the cache and result list
		d.genreCache.add(&genre)
		genres = append(genres, genre)
	}

	return genres, nil
}

// deleteGenresForMovie deletes all genres for the given movie.
func (d *Database) deleteGenresForMovie(movie *Movie) error {
	db, err := d.getDatabase()
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	// Delete all MovieGenre entries with the given movie ID in one step
	if result := db.Where("movie_id = ?", movie.Id).Delete(&MovieGenre{}); result.Error != nil {
		return fmt.Errorf("failed to delete genres for movie ID %d: %w", movie.Id, result.Error)
	}

	return nil
}
