package data

import (
	"fmt"
	"os"
	"path"
	"strings"

	"gorm.io/gorm"
)

// Movie represents a movie in the database.
type Movie struct {
	Id         int      `gorm:"column:id;primary_key"`
	Title      string   `gorm:"column:title;size:100"`
	SubTitle   string   `gorm:"column:sub_title;size:100"`
	Year       int      `gorm:"column:year;"`
	ImdbRating float32  `gorm:"column:imdb_rating;"`
	MyRating   int      `gorm:"column:my_rating;"`
	ImdbUrl    string   `gorm:"column:imdb_url;size:1024"`
	ImdbID     string   `gorm:"column:imdb_id;size:9"`
	StoryLine  string   `gorm:"column:story_line;size:65535"`
	MoviePath  string   `gorm:"column:path;size:1024"`
	Runtime    int      `gorm:"column:length"`
	Genres     []Genre  `gorm:"-"`
	Persons    []Person `gorm:"-"`

	HasImage      bool   `gorm:"-"`
	Image         []byte `gorm:"-"`
	ImageId       int    `gorm:"column:image_id;"`
	ImagePath     string `gorm:"column:image_path;size:1024"` // Not used yet
	ToWatch       bool   `gorm:"column:to_watch"`
	Pack          string `gorm:"column:pack"`
	NeedsSubtitle bool   `gorm:"column:needsSubtitle"`
}

// TableName returns the name of the table.
func (m *Movie) TableName() string {
	return "movies"
}

// SearchMoviesEx returns all movies in the database that matches the search criteria.
func (d *Database) SearchMoviesEx(currentView string, searchFor string, genreId int, orderBy string,
	onlyNotProcessed bool) ([]*Movie, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}

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

	if genreId != -1 && onlyNotProcessed {
		panic("onlyNotProcessed does not work with genre search")
	}

	if genreId == -1 {
		sqlWhere, sqlArgs = getStandardSearch(searchFor, onlyNotProcessed)
	} else {
		sqlJoin, sqlWhere, sqlArgs = getGenreSearch(searchFor, genreId)
	}

	sqlWhere = addViewSQL(currentView, sqlWhere)

	query := db
	if sqlJoin != "" {
		query = query.Joins(sqlJoin)
	}
	if sqlWhere != "" {
		if len(sqlArgs) == 0 {
			query = query.Where(sqlWhere)
		} else {
			query = query.Where(sqlWhere, sqlArgs)
		}
	}
	query = query.Order(sqlOrderBy)
	if onlyNotProcessed {
		query = query.Limit(1)
	}
	if result := query.Find(&movies); result.Error != nil {
		return nil, result.Error
	}

	movies, err = d.getGenresForMovies(movies)
	if err != nil {
		return nil, err
	}

	movies, err = d.getPersonsForMovies(movies)
	if err != nil {
		return nil, err
	}

	movies, err = d.getImagesForMovies(movies)
	if err != nil {
		return nil, err
	}

	return movies, nil
}

// SearchMovies returns all movies in the database that matches the search criteria.
func (d *Database) SearchMovies(currentView string, searchFor string, genreId int, orderBy string) ([]*Movie, error) {
	return d.SearchMoviesEx(currentView, searchFor, genreId, orderBy, false)
}

// GetAllMoviePaths returns a list of all the movie paths in the database. Used when adding new movies.
func (d *Database) GetAllMoviePaths() ([]string, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}

	var movies []*Movie
	if result := db.Find(&movies); result.Error != nil {
		return nil, result.Error
	}

	var paths []string
	for i := range movies {
		paths = append(paths, movies[i].MoviePath)
	}
	return paths, nil
}

// GetAllMovieTitles returns a list of all the movie titles in the database. Used when adding new movies.
func (d *Database) GetAllMovieTitles() ([]string, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}

	var movies []*Movie
	if result := db.Find(&movies); result.Error != nil {
		return nil, result.Error
	}

	var titles []string
	for i := range movies {
		if movies[i].SubTitle != "" {
			titles = append(titles, fmt.Sprintf("%s (%s)", movies[i].Title, movies[i].SubTitle))
		} else {
			titles = append(titles, movies[i].Title)
		}
	}
	return titles, nil
}

// InsertMovie adds a new movie to the database.
func (d *Database) InsertMovie(movie *Movie) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}

	err = db.Transaction(
		func(tx *gorm.DB) error {
			// Insert image
			if movie.HasImage && len(movie.Image) > 0 {
				image := image{Data: movie.Image}
				err = d.insertImage(&image)
				if err != nil {
					return err
				}
				movie.ImageId = image.Id
			}

			if result := db.Create(movie); result.Error != nil {
				return result.Error
			}

			// Handle genres
			for i := range movie.Genres {
				genre, err := d.getOrInsertGenre(&movie.Genres[i])
				if err != nil {
					return err
				}

				err = d.InsertMovieGenre(movie, genre)
				if err != nil {
					return err
				}
			}

			// Handle persons
			for i := range movie.Persons {
				person, err := d.InsertPerson(&movie.Persons[i])
				if err != nil {
					return err
				}

				err = d.InsertMoviePerson(movie, person)
				if err != nil {
					return err
				}
			}

			return nil
		},
	)

	// Check transaction error
	if err != nil {
		return err
	}

	return nil
}

// UpdateMovie update a movie.
func (d *Database) UpdateMovie(movie *Movie, updateGenres bool) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}

	err = db.Transaction(
		func(tx *gorm.DB) error {
			updates := make(map[string]interface{}, 12)

			updates["title"] = movie.Title
			updates["sub_title"] = movie.SubTitle
			updates["story_line"] = movie.StoryLine
			updates["imdb_rating"] = movie.ImdbRating
			updates["imdb_url"] = movie.ImdbUrl
			updates["year"] = movie.Year
			updates["my_rating"] = movie.MyRating
			updates["to_watch"] = movie.ToWatch
			updates["image_id"] = movie.ImageId
			updates["pack"] = movie.Pack
			updates["needsSubtitle"] = movie.NeedsSubtitle
			updates["length"] = movie.Runtime

			if result := db.Model(&movie).Updates(updates); result.Error != nil {
				return result.Error
			}

			if !updateGenres {
				return nil
			}

			// Handle genres
			for i := range movie.Genres {
				genre, err := d.getOrInsertGenre(&movie.Genres[i])
				if err != nil {
					return err
				}

				err = d.RemoveMovieGenre(movie, genre)
				if err != nil {
					return err
				}

				err = d.InsertMovieGenre(movie, genre)
				if err != nil {
					return err
				}
			}

			return nil
		},
	)

	// Check transaction error
	if err != nil {
		return err
	}

	return nil
}

// UpdateMoviePersons update a movie with its directors, writers and actors.
func (d *Database) UpdateMoviePersons(movie *Movie) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}

	err = db.Transaction(
		func(tx *gorm.DB) error {
			// Handle persons
			for i := range movie.Persons {
				person, err := d.GetPerson(movie.Persons[i].Name)
				if err != nil {
					return err
				}
				if person == nil {
					person, err = d.InsertPerson(&movie.Persons[i])
					if err != nil {
						return err
					}
				}

				person.Type = movie.Persons[i].Type

				err = d.InsertMoviePerson(movie, person)
				if err != nil {
					return err
				}
			}

			return nil
		},
	)

	// Check transaction error
	if err != nil {
		return err
	}

	return nil
}

// DeleteMovie removes a movie from the database.
func (d *Database) DeleteMovie(rootDir string, movie *Movie) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}

	err = db.Transaction(
		func(tx *gorm.DB) error {
			if err = d.deleteImage(movie); err != nil {
				return err
			}

			if err = d.deleteGenresForMovie(movie); err != nil {
				return err
			}

			if err = d.deletePersonsForMovie(movie); err != nil {
				return err
			}

			if result := db.Delete(movie, movie.Id); result.Error != nil {
				return result.Error
			}

			return nil
		},
	)

	if err != nil {
		return err
	}

	moviePath := path.Join(rootDir, movie.MoviePath)
	err = os.RemoveAll(moviePath)
	if err != nil {
		return err
	}

	return nil
}

func addViewSQL(view string, where string) string {
	var sql string
	switch view {
	case "packs":
		sql = "pack != '' AND pack is not null"
	case "toWatch":
		sql = "to_watch = true AND needsSubtitle = false"
	case "noRating":
		sql = "my_rating = 0 AND needsSubtitle = false"
	case "needsSubtitles":
		sql = "needsSubtitle = true"
	}
	if where != "" && sql != "" {
		sql = " AND " + sql
	}
	if where != "" {
		where = "(" + where + ")"
	}
	return where + sql
}

func getGenreSearch(searchFor string, genreId int) (string, string, map[string]interface{}) {
	var sqlWhere, sqlJoin string
	var sqlArgs map[string]interface{}
	sqlArgs = make(map[string]interface{})

	if searchFor == "" {
		sqlJoin = "JOIN movie_genre on movies.id = movie_genre.movie_id"
		sqlWhere = "movie_genre.genre_id = @genre"
		sqlArgs["genre"] = genreId
	} else {
		sqlJoin = "JOIN movie_genre on movies.id = movie_genre.movie_id"
		sqlWhere = "(title like @search OR sub_title like @search OR year like @search OR story_line like @search" +
			") AND movie_genre.genre_id = @genre"
		sqlArgs["search"] = "%" + searchFor + "%"
		sqlArgs["genre"] = genreId
	}
	return sqlJoin, sqlWhere, sqlArgs
}

func getStandardSearch(searchFor string, onlyNotProcessed bool) (string, map[string]interface{}) {
	var sqlWhere string
	var sqlArgs map[string]interface{}
	sqlArgs = make(map[string]interface{})
	if searchFor != "" {
		prefix, search := getSearchPrefix(searchFor)
		switch prefix {
		case "title":
			sqlWhere = "title like @search OR sub_title like @search"
		case "year":
			sqlWhere = "year like @search"
		case "pack":
			sqlWhere = "pack like @search"
		case "imdb":
			sqlWhere = "imdb_rating >= @search"
		case "myrating":
			sqlWhere = "my_rating >= @search"
		default:
			sqlWhere = "title like @search OR sub_title like @search OR year like @search OR story_line like @search"
		}
		sqlArgs["search"] = search
	}

	if onlyNotProcessed {
		if sqlWhere != "" {
			sqlWhere = "processed=false && (" + sqlWhere + ")"
		} else {
			sqlWhere = "processed=false"
		}
	}

	return sqlWhere, sqlArgs
}

func getSearchPrefix(searchFor string) (string, string) {
	before, after, _ := strings.Cut(searchFor, ":")
	switch before {
	case "title", "pack":
		return before, "%" + after + "%"
	case "year", "imdb", "myrating":
		return before, after
	default:
		return "", "%" + searchFor + "%"
	}
}

func (d *Database) getImagesForMovies(movies []*Movie) ([]*Movie, error) {
	// Get images for movies
	for i := range movies {
		movie := movies[i]

		// Check cache for image
		img := d.cache.load(movie.Id)
		if img != nil {
			movie.Image = img
			continue
		}

		// Image is not in cache, so load it from database
		// and store it in cache
		d.getMovieImage(movie)
		d.cache.save(movie.Id, movie.Image)
	}
	return movies, nil
}

func (d *Database) getGenresForMovies(movies []*Movie) ([]*Movie, error) {
	// Get images for movies
	for i := range movies {
		movie := movies[i]

		// Load genres
		genres, err := d.getGenresForMovie(movie)
		if err != nil {
			return nil, err
		}
		movie.Genres = genres
	}
	return movies, nil
}

func (d *Database) getPersonsForMovies(movies []*Movie) ([]*Movie, error) {
	// Get images for movies
	for i := range movies {
		movie := movies[i]

		// Load persons
		persons, err := d.GetPersonsForMovie(movie)
		if err != nil {
			return nil, err
		}
		movie.Persons = persons
	}
	return movies, nil
}

func (d *Database) getMovieImage(movie *Movie) {

	// Get image (if it exists)
	if movie.ImageId > 0 {
		img, err := d.getImage(movie.ImageId)
		if err != nil {
			return
		}
		movie.Image = img.Data
		movie.HasImage = true
	}
}

// SetProcessed sets the movie as processed
func (d *Database) SetProcessed(movie *Movie) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}

	updates := make(map[string]interface{}, 1)

	updates["processed"] = true

	if result := db.Model(&movie).Updates(updates); result.Error != nil {
		return result.Error
	}

	return nil
}
