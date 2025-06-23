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
	Id int `gorm:"column:id;primary_key"`

	Title     string   `gorm:"column:title;size:100"`
	SubTitle  string   `gorm:"column:sub_title;size:100"`
	StoryLine string   `gorm:"column:story_line;size:65535"`
	Year      int      `gorm:"column:year;"`
	MyRating  int      `gorm:"column:my_rating;"`
	MoviePath string   `gorm:"column:path;size:1024"`
	Runtime   int      `gorm:"column:length"`
	Genres    []Genre  `gorm:"-"`
	Persons   []Person `gorm:"-"`

	ImdbRating float32 `gorm:"column:imdb_rating;"`
	ImdbUrl    string  `gorm:"column:imdb_url;size:1024"`
	ImdbID     string  `gorm:"column:imdb_id;size:9"`

	HasImage bool   `gorm:"-"`
	Image    []byte `gorm:"-"`
	ImageId  int    `gorm:"column:image_id;"`

	ToWatch       bool   `gorm:"column:to_watch"`
	Pack          string `gorm:"column:pack"`
	NeedsSubtitle bool   `gorm:"column:needsSubtitle"`
}

var personType = map[string]int{
	"person":   -1,
	"director": 0,
	"writer":   1,
	"actor":    2,
}

// TableName returns the name of the table.
func (m *Movie) TableName() string {
	return "movies"
}

// SearchMovies returns all movies in the database that matches the search criteria.
func (d *Database) SearchMovies(currentView string, searchFor string, genreId int, orderBy string) ([]*Movie, error) {
	return d.SearchMoviesEx(currentView, searchFor, genreId, orderBy)
}

// SearchMoviesEx returns all movies in the database that matches the search criteria.
func (d *Database) SearchMoviesEx(currentView string, searchFor string, genreId int, orderBy string) ([]*Movie, error) {

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

	parts := strings.Split(searchFor, ":")
	typ, ok := personType[parts[0]]
	if ok {
		searchFor = searchFor[len(parts[0])+1:]
		sqlJoin, sqlWhere, sqlArgs = getPersonSearch(searchFor, typ)
	} else if genreId == -1 {
		sqlWhere, sqlArgs = getStandardSearch(searchFor)
	} else {
		sqlJoin, sqlWhere, sqlArgs = getGenreSearch(searchFor, genreId)
	}

	sqlWhere = addViewSQL(currentView, sqlWhere)

	query, err := d.getQuery(sqlJoin, sqlWhere, sqlArgs, sqlOrderBy)
	if err != nil {
		return nil, fmt.Errorf("failed to get query : %w", err)
	}

	if err := query.Distinct().Find(&movies).Error; err != nil {
		return nil, fmt.Errorf("failed to find movies: %w", err)
	}

	movies, err = d.getGenresForMovies(movies)
	if err != nil {
		return nil, fmt.Errorf("failed to get genres for movies: %w", err)
	}

	movies, err = d.getImagesForMovies(movies)
	if err != nil {
		return nil, fmt.Errorf("failed to get images for movies: %w", err)
	}

	return movies, nil
}

func (d *Database) getQuery(sqlJoin string, sqlWhere string, sqlArgs map[string]interface{}, sqlOrderBy string) (*gorm.DB, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	if sqlJoin != "" {
		db = db.Joins(sqlJoin)
	}
	if sqlWhere != "" {
		if len(sqlArgs) == 0 {
			db = db.Where(sqlWhere)
		} else {
			db = db.Where(sqlWhere, sqlArgs)
		}
	}

	db = db.Order(sqlOrderBy)

	return db, nil
}

// GetAllMoviePaths returns a list of all the movie paths in the database. Used when adding new movies.
func (d *Database) GetAllMoviePaths() ([]string, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	var movies []*Movie
	if err := db.Find(&movies).Error; err != nil {
		return nil, fmt.Errorf("failed to get paths for movies: %w", err)
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
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	var movies []*Movie
	if err := db.Find(&movies).Error; err != nil {
		return nil, fmt.Errorf("failed to get movie titles: %w", err)
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
		return fmt.Errorf("failed to get database: %w", err)
	}

	err = db.Transaction(
		func(tx *gorm.DB) error {
			// Insert image
			if movie.HasImage && len(movie.Image) > 0 {
				image := image{Data: movie.Image}
				err = d.createImage(&image)
				if err != nil {
					return fmt.Errorf("failed to create image: %w", err)
				}
				movie.ImageId = image.Id
			}

			if err := db.Create(movie).Error; err != nil {
				return fmt.Errorf("failed to create movie: %w", err)
			}

			// Handle genres
			for i := range movie.Genres {
				genre, err := d.getOrInsertGenre(&movie.Genres[i])
				if err != nil {
					return fmt.Errorf("failed to get or create movie genre: %w", err)
				}

				err = d.InsertMovieGenre(movie, genre)
				if err != nil {
					return fmt.Errorf("failed to insert movie genre id: %w", err)
				}
			}

			// Handle persons
			for _, person := range movie.Persons {
				// To prevent a problem with the type getting overwritten
				// with the zero value (0 = Director) for existing persons
				t := person.Type

				p, err := d.GetPerson(person.Name)
				if err != nil {
					return fmt.Errorf("failed to get person: %w", err)
				}

				if p == nil {
					p, err = d.InsertPerson(&person)
					if err != nil {
						return fmt.Errorf("failed to insert person: %w", err)
					}
				}

				p.Type = t

				err = d.InsertMoviePerson(movie, p)
				if err != nil {
					return fmt.Errorf("failed to update movie person id: %w", err)
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
func (d *Database) UpdateMovie(movie *Movie) error {
	db, err := d.getDatabase()
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
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

			if err := db.Model(&movie).Updates(updates).Error; err != nil {
				return fmt.Errorf("failed to update movie: %w", err)
			}

			// Handle genres
			for i := range movie.Genres {
				genre, err := d.getOrInsertGenre(&movie.Genres[i])
				if err != nil {
					return fmt.Errorf("failed to get or insert movie genre: %w", err)
				}

				err = d.getOrInsertMovieGenre(movie, genre)
				if err != nil {
					return fmt.Errorf("failed to update movie genre id: %w", err)
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
		return fmt.Errorf("failed to get database: %w", err)
	}

	err = db.Transaction(
		func(tx *gorm.DB) error {
			// Handle persons
			for i := range movie.Persons {
				person, err := d.GetPerson(movie.Persons[i].Name)
				if err != nil {
					return fmt.Errorf("failed to get person: %w", err)
				}
				if person == nil {
					person, err = d.InsertPerson(&movie.Persons[i])
					if err != nil {
						return fmt.Errorf("failed to insert person: %w", err)
					}
				}

				person.Type = movie.Persons[i].Type

				err = d.InsertMoviePerson(movie, person)
				if err != nil {
					return fmt.Errorf("failed to update movie person id: %w", err)
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
		return fmt.Errorf("failed to get database: %w", err)
	}

	err = db.Transaction(
		func(tx *gorm.DB) error {
			if err = d.deleteImage(movie); err != nil {
				return fmt.Errorf("failed to delete movie image: %w", err)
			}

			if err = d.deleteGenresForMovie(movie); err != nil {
				return fmt.Errorf("failed to delete movie genres: %w", err)
			}

			if err = d.deletePersonsForMovie(movie); err != nil {
				return fmt.Errorf("failed to delete movie persons: %w", err)
			}

			if result := db.Delete(movie, movie.Id); result.Error != nil {
				return fmt.Errorf("failed to delete movie: %w", result.Error)
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

// addViewSQL returns a combined SQL WHERE clause based on the given view and optional base clause.
func addViewSQL(view, baseWhere string) string {
	viewConditions := map[string]string{
		"packs":          "pack != '' AND pack IS NOT NULL",
		"toWatch":        "to_watch = true AND needsSubtitle = false",
		"noRating":       "my_rating = 0 AND needsSubtitle = false",
		"needsSubtitles": "needsSubtitle = true",
	}

	viewWhere, ok := viewConditions[view]
	if !ok {
		viewWhere = ""
	}

	var clauses []string

	if baseWhere != "" {
		clauses = append(clauses, "("+baseWhere+")")
	}
	if viewWhere != "" {
		clauses = append(clauses, viewWhere)
	}

	return strings.Join(clauses, " AND ")
}

func getGenreSearch(searchFor string, genreId int) (join string, where string, args map[string]interface{}) {
	join = "JOIN movie_genre ON movies.id = movie_genre.movie_id"
	args = map[string]interface{}{
		"genre": genreId,
	}

	if searchFor == "" {
		where = "movie_genre.genre_id = @genre"
	} else {
		where = `(title LIKE @search OR sub_title LIKE @search OR year LIKE @search OR story_line LIKE @search)
		         AND movie_genre.genre_id = @genre`
		args["search"] = "%" + searchFor + "%"
	}

	return
}

func getPersonSearch(searchFor string, typ int) (join string, where string, args map[string]interface{}) {
	join = `
		JOIN movie_person ON movies.id = movie_person.movie_id
		JOIN person ON person.id = movie_person.person_id
	`

	where = "person.name LIKE @search"
	args = map[string]interface{}{
		"search": "%" + searchFor + "%",
	}

	if typ >= 0 && typ <= 2 {
		where += " AND movie_person.type = @type"
		args["type"] = typ
	}

	return
}

func getStandardSearch(searchFor string) (string, map[string]interface{}) {
	sqlArgs := make(map[string]interface{})
	var conditions []string

	if searchFor != "" {
		prefix, search := getSearchPrefix(searchFor)
		var condition string

		switch prefix {
		case "title":
			condition = "title LIKE @search OR sub_title LIKE @search"
		case "year":
			condition = "year LIKE @search"
		case "pack":
			condition = "pack LIKE @search"
		case "imdb":
			condition = "imdb_rating >= @search"
		case "myrating":
			condition = "my_rating >= @search"
		default:
			condition = "title LIKE @search OR sub_title LIKE @search OR year LIKE @search OR story_line LIKE @search"
		}

		conditions = append(conditions, "("+condition+")")
		sqlArgs["search"] = search
	}

	sqlWhere := strings.Join(conditions, " AND ")

	return sqlWhere, sqlArgs
}

func getSearchPrefix(searchFor string) (string, string) {
	before, after, _ := strings.Cut(searchFor, ":")
	before = strings.TrimSpace(before)
	after = strings.TrimSpace(after)

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
		img := d.imageCache.load(movie.Id)
		if img != nil {
			movie.Image = img
			continue
		}

		// The image is not in the cache, so load it from the database
		// and store it in the cache
		d.getImageForMovie(movie)
		d.imageCache.save(movie.Id, movie.Image)
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
			return nil, fmt.Errorf("failed to get genres for movie (%d: %s): %w", movie.Id, movie.Title, err)
		}
		movie.Genres = genres
	}
	return movies, nil
}

func (d *Database) GetPersonsForMovies(movies []*Movie) ([]*Movie, error) {
	// Get images for movies
	for i := range movies {
		movie := movies[i]

		// Load persons
		persons, err := d.GetPersonsForMovie(movie)
		if err != nil {
			return nil, fmt.Errorf("failed to get person for movie (%d: %s): %w", movie.Id, movie.Title, err)
		}
		movie.Persons = persons
	}
	return movies, nil
}

func (d *Database) getImageForMovie(movie *Movie) {

	// Get image (if it exists)
	if movie.ImageId > 0 {
		img, err := d.readImage(movie.ImageId)
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
		return fmt.Errorf("failed to get database: %w", err)
	}

	updates := make(map[string]interface{}, 1)

	updates["processed"] = true

	if err := db.Model(&movie).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to movie as processed: %w", err)
	}

	return nil
}
