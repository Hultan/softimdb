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

	movieGenre := MovieGenre{MovieId: movie.Id, GenreId: Genre.Id}

	// If it does not, create it
	if result := db.Create(movieGenre); result.Error != nil {
		return result.Error
	}
	return nil
}

// RemoveMovieGenre removes a movie genre from the database.
func (d *Database) RemoveMovieGenre(movie *Movie, genre *Genre) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("delete from movie_genre where movie_id = %v and genre_id = %v", movie.Id, genre.Id)
	tx := db.Exec(sql)

	if tx.Error != nil {
		return tx.Error
	}

	return nil
}
