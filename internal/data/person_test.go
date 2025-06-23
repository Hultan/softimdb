package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabase_Person(t *testing.T) {
	db, err := openDatabase()
	if err != nil {
		t.Fatal(err)
	}
	defer db.CloseDatabase()

	person := &Person{
		Name: "John Doe",
	}

	got, err := db.GetPerson(person.Name)
	if err != nil {
		t.Errorf("GetPerson() error = %v", err)
		return
	}

	if got == nil {
		got, err = db.InsertPerson(person)
		if err != nil {
			t.Errorf("InsertPerson() error = %v", err)
			return
		}
	}

	assert.GreaterOrEqual(t, got.Id, 0, "person id should not be zero")

	got, err = db.GetPerson(person.Name)
	if err != nil {
		t.Errorf("GetPerson() error = %v", err)
	}

	assert.Equal(t, person.Name, got.Name, "person name should be John Doe")

	movie, err := getMovie(db)
	if err != nil {
		t.Errorf("getMovie() error = %v", err)
	}

	err = db.InsertMoviePerson(movie, got)
	if err != nil {
		t.Fatal(err)
	}

	movie, err = getMovie(db)
	if err != nil {
		t.Errorf("getMovie() error = %v", err)
	}

	assert.Equal(t, 1, len(movie.Persons), "number of persons should be 1")
	assert.Equal(t, person.Name, movie.Persons[0].Name, "person name should be John Doe")
	assert.Equal(t, person.Type, movie.Persons[0].Type, "person type should be Writer")

	err = db.RemovePerson(person)
	if err != nil {
		t.Errorf("RemovePerson() error = %v", err)
	}
}

func getMovie(db *Database) (*Movie, error) {
	movies, err := db.SearchMovies("all", "gladiator", -1, "title")
	if err != nil {
		return nil, err
	}

	return movies[0], nil
}
