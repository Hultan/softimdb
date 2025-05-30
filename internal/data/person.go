package data

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strings"

	"gorm.io/gorm"
)

type PersonType int

const (
	Director PersonType = iota
	Writer
	Actor
)

// Person represents a person (director, writer or actor).
type Person struct {
	Id     int        `gorm:"column:id;primary_key"`
	Name   string     `gorm:"column:name;size:50"`
	Type   PersonType `gorm:"-"`
	Movies []Movie    `gorm:"-"`
}

// TableName returns the person's table name.
func (m *Person) TableName() string {
	return "person"
}

// GetPerson returns a person by name.
func (d *Database) GetPerson(name string) (*Person, error) {
	name = strings.TrimSpace(name)

	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}

	person := Person{}
	if err := db.Where("name=?", name).First(&person).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &person, nil
}

// InsertPerson inserts a new person and returns it.
func (d *Database) InsertPerson(person *Person) (*Person, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}

	person.Name = strings.TrimSpace(person.Name)

	if err := db.Create(person).Error; err != nil {
		return nil, err

	}

	return person, nil
}

// GetPersonsForMovie returns a list of persons (director, writer or actor) connected to the given movie.
func (d *Database) GetPersonsForMovie(movie *Movie) ([]Person, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}

	// Get person id:s for the movie
	var moviePersons []MoviePerson
	if err := db.Where("movie_id=?", movie.Id).Find(&moviePersons).Error; err != nil {
		return nil, err
	}

	var persons []Person

	for i := range moviePersons {
		var person Person
		if err := db.Where("id=?", moviePersons[i].PersonId).Find(&person).Error; err != nil {
			return nil, err
		}
		person.Type = PersonType(moviePersons[i].Type)
		persons = append(persons, person)
	}

	slices.SortFunc(persons, func(a, b Person) int {
		return cmp.Compare(a.Type, b.Type)
	})

	return persons, nil
}

// RemovePerson removes a person from the database, including associations.
func (d *Database) RemovePerson(person *Person) error {
	db, err := d.getDatabase()
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	// Use parameterized queries for safety
	if err := db.Exec("DELETE FROM movie_person WHERE person_id = ?", person.Id).Error; err != nil {
		return fmt.Errorf("failed to delete movie_person entries: %w", err)
	}

	if err := db.Exec("DELETE FROM person WHERE id = ?", person.Id).Error; err != nil {
		return fmt.Errorf("failed to delete person: %w", err)
	}

	return nil
}

// deletePersonsForMovie removes all person associations for the given movie.
func (d *Database) deletePersonsForMovie(movie *Movie) error {
	db, err := d.getDatabase()
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	if err := db.Exec("DELETE FROM movie_person WHERE movie_id = ?", movie.Id).Error; err != nil {
		return fmt.Errorf("failed to delete movie_person entries for movie ID %d: %w", movie.Id, err)
	}

	return nil
}
