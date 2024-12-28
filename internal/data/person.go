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

// TableName returns the person table name.
func (m *Person) TableName() string {
	return "person"
}

// GetPerson returns a person by name.
func (d *Database) GetPerson(name string) (*Person, error) {
	name = strings.Trim(name, " \t\n")

	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}

	person := Person{}
	if result := db.Where("name=?", name).First(&person); result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
		return nil, nil
	}

	return &person, nil
}

// InsertPerson inserts a new person and returns it.
func (d *Database) InsertPerson(person *Person) (*Person, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}

	person.Name = strings.Trim(person.Name, " \t\n")

	// If it does not, create it
	if result := db.Create(person); result.Error != nil {
		return nil, result.Error
	}
	return person, nil
}

// GetPersonsForMovie returns a list of persons (director, writer or actor) connected to the given movie.
func (d *Database) GetPersonsForMovie(movie *Movie) ([]Person, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}

	// Get person id:s for movie
	var moviePersons []MoviePerson
	if result := db.Where("movie_id=?", movie.Id).Find(&moviePersons); result.Error != nil {
		return nil, result.Error
	}

	var persons []Person

	for i := range moviePersons {
		var person Person
		if result := db.Where("id=?", moviePersons[i].PersonId).Find(&person); result.Error != nil {
			return nil, result.Error
		}
		person.Type = PersonType(moviePersons[i].Type)
		persons = append(persons, person)
	}

	slices.SortFunc(persons, func(a, b Person) int {
		return cmp.Compare(a.Type, b.Type)
	})

	return persons, nil
}

// RemovePerson removes a person from the database.
func (d *Database) RemovePerson(person *Person) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("delete from movie_person where person_id = %v", person.Id)
	tx := db.Exec(sql)
	if tx.Error != nil {
		return tx.Error
	}

	sql = fmt.Sprintf("delete from person where id = %v", person.Id)
	tx = db.Exec(sql)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}

// deletePersonsForMovie removes all persons for a movie from the database.
func (d *Database) deletePersonsForMovie(movie *Movie) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}
	sql := fmt.Sprintf("delete from movie_person where movie_id = %v", movie.Id)
	tx := db.Exec(sql)
	if tx.Error != nil {
		return tx.Error
	}

	return nil
}
