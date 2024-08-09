package data

import (
	"os"
	"path"
	"strings"

	"gorm.io/gorm"
)

// Movie represents a movie in the database.
type Movie struct {
	Id         int     `gorm:"column:id;primary_key"`
	Title      string  `gorm:"column:title;size:100"`
	SubTitle   string  `gorm:"column:sub_title;size:100"`
	Year       int     `gorm:"column:year;"`
	ImdbRating float32 `gorm:"column:imdb_rating;"`
	MyRating   int     `gorm:"column:my_rating;"`
	ImdbUrl    string  `gorm:"column:imdb_url;size:1024"`
	ImdbID     string  `gorm:"column:imdb_id;size:9"`
	StoryLine  string  `gorm:"column:story_line;size:65535"`
	MoviePath  string  `gorm:"column:path;size:1024"`
	Runtime    int     `gorm:"column:length"`
	Tags       []Tag   `gorm:"-"`

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

// GetMovie returns a movie from the database.
func (d *Database) GetMovie(id int) (*Movie, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}
	movie := Movie{}
	if result := db.Where("Id=?", id).First(&movie); result.Error != nil {
		return nil, result.Error
	}

	// Check cache
	image := d.cache.Load(movie.Id)
	if image != nil {
		movie.Image = image
	} else {
		// Get movie image
		d.getMovieImage(&movie)
	}

	// Get tags for movie
	tags, err := d.GetTagsForMovie(&movie)
	if err != nil {
		return nil, err
	}
	movie.Tags = tags

	return &movie, nil
}

// GetAllMovies returns all movies in the database that matches the search criteria.
func (d *Database) GetAllMovies(currentView string, searchFor string, categoryId int, orderBy string) ([]*Movie, error) {
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

	if categoryId == -1 {
		sqlWhere, sqlArgs = getStandardSearch(searchFor)
	} else {
		sqlJoin, sqlWhere, sqlArgs = getCategorySearch(searchFor, categoryId)
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
	if result := query.Order(sqlOrderBy).Find(&movies); result.Error != nil {
		return nil, result.Error
	}

	movies, err = d.getImagesForMovies(movies)
	if err != nil {
		return nil, err
	}

	return movies, nil
}

func addViewSQL(view string, where string) string {
	var sql string
	switch view {
	case "packs":
		sql = "pack != '' AND pack is not null"
	case "toWatch":
		sql = "to_watch = true"
	case "noRating":
		sql = "my_rating = 0"
	}
	if where != "" && sql != "" {
		sql = " AND " + sql
	}
	if where != "" {
		where = "(" + where + ")"
	}
	return where + sql
}

func getCategorySearch(searchFor string, categoryId int) (string, string, map[string]interface{}) {
	var sqlWhere, sqlJoin string
	var sqlArgs map[string]interface{}
	sqlArgs = make(map[string]interface{})

	if searchFor == "" {
		sqlJoin = "JOIN movie_tag on movies.id = movie_tag.movie_id"
		sqlWhere = "movie_tag.tag_id = @category"
		sqlArgs["category"] = categoryId
	} else {
		sqlJoin = "JOIN movie_tag on movies.id = movie_tag.movie_id"
		sqlWhere = "(title like @search OR sub_title like @search OR year like @search OR story_line like @search) AND movie_tag.tag_id = @category"
		sqlArgs["search"] = "%" + searchFor + "%"
		sqlArgs["category"] = categoryId
	}
	return sqlJoin, sqlWhere, sqlArgs
}

func getStandardSearch(searchFor string) (string, map[string]interface{}) {
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

		// Load genres (tags)
		tags, err := d.GetTagsForMovie(movie)
		if err != nil {
			return nil, err
		}
		movie.Tags = tags

		// Check cache for image
		image := d.cache.Load(movie.Id)
		if image != nil {
			movie.Image = image
			continue
		}

		// Image is not in cache, so load it from database
		// and store it in cache
		d.getMovieImage(movie)
		d.cache.Save(movie.Id, movie.Image)
	}
	return movies, nil
}

// GetAllMoviePaths returns a list of all the movie paths in the database. Used when adding new movies.
func (d *Database) GetAllMoviePaths() (*[]string, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}

	var movies []*Movie
	if result := db.Find(&movies); result.Error != nil {
		return nil, result.Error
	}

	var paths = &[]string{}
	for i := range movies {
		*paths = append(*paths, movies[i].MoviePath)
	}
	return paths, nil
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
				image := Image{Data: movie.Image}
				err = d.InsertImage(&image)
				if err != nil {
					return err
				}
				movie.ImageId = image.Id
			}

			if result := db.Create(movie); result.Error != nil {
				return result.Error
			}

			// Handle tags
			for i := range movie.Tags {
				tag, err := d.GetOrInsertTag(&movie.Tags[i])
				if err != nil {
					return err
				}

				err = d.InsertMovieTag(movie, tag)
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
func (d *Database) UpdateMovie(movie *Movie, updateTags bool) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}

	err = db.Transaction(
		func(tx *gorm.DB) error {
			if result := db.Model(&movie).Update("title", movie.Title); result.Error != nil {
				return result.Error
			}

			if result := db.Model(&movie).Update("sub_title", movie.SubTitle); result.Error != nil {
				return result.Error
			}

			if result := db.Model(&movie).Update("story_line", movie.StoryLine); result.Error != nil {
				return result.Error
			}

			if result := db.Model(&movie).Update("imdb_rating", movie.ImdbRating); result.Error != nil {
				return result.Error
			}

			if result := db.Model(&movie).Update("imdb_url", movie.ImdbUrl); result.Error != nil {
				return result.Error
			}

			if result := db.Model(&movie).Update("year", movie.Year); result.Error != nil {
				return result.Error
			}

			if result := db.Model(&movie).Update("my_rating", movie.MyRating); result.Error != nil {
				return result.Error
			}

			if result := db.Model(&movie).Update("to_watch", movie.ToWatch); result.Error != nil {
				return result.Error
			}

			if result := db.Model(&movie).Update("image_id", movie.ImageId); result.Error != nil {
				return result.Error
			}

			if result := db.Model(&movie).Update("pack", movie.Pack); result.Error != nil {
				return result.Error
			}

			if result := db.Model(&movie).Update("needsSubtitle", movie.NeedsSubtitle); result.Error != nil {
				return result.Error
			}

			if result := db.Model(&movie).Update("length", movie.Runtime); result.Error != nil {
				return result.Error
			}

			if !updateTags {
				return nil
			}

			// Handle tags
			for i := range movie.Tags {
				tag, err := d.GetOrInsertTag(&movie.Tags[i])
				if err != nil {
					return err
				}

				err = d.RemoveMovieTag(movie, tag)
				if err != nil {
					return err
				}

				err = d.InsertMovieTag(movie, tag)
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
			if err = d.DeleteImage(movie); err != nil {
				return err
			}

			if err = d.DeleteTagsForMovie(movie); err != nil {
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

func (d *Database) getMovieImage(movie *Movie) {

	// Get image (if it exists)
	if movie.ImageId > 0 {
		image, err := d.GetImage(movie.ImageId)
		if err != nil {
			return
		}
		movie.Image = image.Data
		movie.HasImage = true
	}
}
