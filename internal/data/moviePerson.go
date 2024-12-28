package data

// MoviePerson represents a person (director, writer or actor) and a movie.
type MoviePerson struct {
	MovieId  int `gorm:"column:movie_id;primary_key;"`
	PersonId int `gorm:"column:person_id;primary_key;"`
	Type     int `gorm:"column:type;"`
}

// TableName returns the person table name.
func (m *MoviePerson) TableName() string {
	return "movie_person"
}

// InsertMoviePerson inserts a person (director, writer or actor) into the database.
func (d *Database) InsertMoviePerson(movie *Movie, person *Person) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}

	moviePerson := MoviePerson{
		MovieId:  movie.Id,
		PersonId: person.Id,
		Type:     int(person.Type),
	}

	if result := db.FirstOrCreate(&moviePerson); result.Error != nil {
		return result.Error
	}
	return nil
}
