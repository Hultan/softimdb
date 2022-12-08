package data

import (
	"fmt"
)

// MovieTag represents a tag and a movie.
type MovieTag struct {
	MovieId int `gorm:"column:movie_id;primary_key;"`
	TagId   int `gorm:"column:tag_id;primary_key;"`
}

// TableName returns the movie_tag table name.
func (m *MovieTag) TableName() string {
	return "movie_tag"
}

// InsertMovieTag inserts a movie tag into the database.
func (d *Database) InsertMovieTag(movie *Movie, tag *Tag) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}

	movieTag := MovieTag{MovieId: movie.Id, TagId: tag.Id}

	// If it does not, create it
	if result := db.Create(movieTag); result.Error != nil {
		return result.Error
	}
	return nil
}

// RemoveMovieTag removes a movie tag from the database.
func (d *Database) RemoveMovieTag(movie *Movie, tag *Tag) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("delete from movie_tag where movie_id = %v and tag_id = %v", movie.Id, tag.Id)
	tx := db.Exec(sql)

	if tx.Error != nil {
		return tx.Error
	}

	return nil
}
