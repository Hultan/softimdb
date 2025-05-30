package data

import (
	"fmt"
)

// MovieGenre represents a Genre and a movie.
type MovieGenre struct {
	MovieId int `gorm:"column:movie_id;primary_key;"`
	GenreId int `gorm:"column:genre_id;primary_key;"`
}

// TableName returns the movie_Genre table name.
func (m *MovieGenre) TableName() string {
	return "movie_genre"
}

// InsertMovieGenre inserts a movie Genre into the database.
func (d *Database) InsertMovieGenre(movie *Movie, Genre *Genre) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}

	movieGenre := MovieGenre{
		MovieId: movie.Id,
		GenreId: Genre.Id,
	}

	if err := db.Create(movieGenre).Error; err != nil {
		return err
	}

	return nil
}

// RemoveMovieGenre removes a movie genre from the database.
// RemoveMovieGenre removes a genre association from a movie.
func (d *Database) RemoveMovieGenre(movie *Movie, genre *Genre) error {
	db, err := d.getDatabase()
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	if err := db.Exec("DELETE FROM movie_genre WHERE movie_id = ? AND genre_id = ?", movie.Id, genre.Id).Error; err != nil {
		return fmt.Errorf("failed to delete movie_genre for movie ID %d and genre ID %d: %w", movie.Id, genre.Id, err)
	}

	return nil
}

// getOrInsertMovieGenre creates a movie_genre record if it does not exist.
func (d *Database) getOrInsertMovieGenre(movie *Movie, genre *Genre) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}

	movieGenre := MovieGenre{
		MovieId: movie.Id,
		GenreId: genre.Id,
	}

	if err := db.FirstOrCreate(&movieGenre).Error; err != nil {
		return err
	}

	return nil
}
