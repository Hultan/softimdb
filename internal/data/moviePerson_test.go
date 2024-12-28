package data

import (
	"testing"

	"github.com/hultan/softimdb/internal/config"
)

func openDatabase() (*Database, error) {
	const configFile = "/home/per/.config/softteam/softimdb/config.json"
	cnf, err := config.LoadConfig(configFile)
	if err != nil {
		return nil, err
	}
	return DatabaseNew(false, cnf), nil
}

func TestDatabase_InsertMoviePerson(t *testing.T) {
	d, err := openDatabase()
	if err != nil {
		t.Fatal(err)
	}

	movie, err := getMovie(d)
	if err != nil {
		t.Fatal(err)
	}

	person := &Person{
		Name: "Test person (REMOVE)",
		Type: 0,
	}

	if err := d.InsertMoviePerson(movie, person); err != nil {
		t.Errorf("InsertMoviePerson() error = %v", err)
	}

	d.CloseDatabase()
}
