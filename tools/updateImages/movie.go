package main

import (
	"fmt"
)

// Movie represents a movie in the database.
type Movie struct {
	Id       int    `gorm:"column:id;primary_key"`
	Title    string `gorm:"column:title;size:100"`
	HasImage bool   `gorm:"-"`
	Image    []byte `gorm:"-"`
	ImageId  int    `gorm:"column:image_id;"`
	URL      string `gorm:"column:imdb_url;size:255"`
}

// TableName returns the name of the table.
func (m *Movie) TableName() string {
	return "movies"
}

// SearchMovies returns all movies in the database that matches the search criteria.
func (d *Database) SearchMovies(currentView string, searchFor string, genreId int, orderBy string) ([]*Movie, error) {
	var (
		movies                        []*Movie
		sqlJoin, sqlWhere, sqlOrderBy string
		sqlArgs                       map[string]interface{}
	)

	if currentView == "packs" && orderBy == "title asc" {
		sqlOrderBy = "pack asc, " + orderBy
	} else {
		sqlOrderBy = orderBy
	}

	sqlWhere, sqlArgs = "", nil

	query, err := d.getQuery(sqlJoin, sqlWhere, sqlArgs, sqlOrderBy)
	if err != nil {
		return nil, fmt.Errorf("failed to get query : %w", err)
	}

	if err := query.Distinct().Find(&movies).Error; err != nil {
		return nil, fmt.Errorf("failed to find movies: %w", err)
	}

	return movies, nil
}
