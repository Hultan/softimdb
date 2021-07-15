package data

import "gorm.io/gorm"

type Movie struct {
	Id         int     `gorm:"column:id;primary_key"`
	Title      string  `gorm:"column:title;size:100"`
	SubTitle   string  `gorm:"column:sub_title;size:100"`
	Year       int     `gorm:"column:year;"`
	ImdbRating float32 `gorm:"column:imdb_rating;"`
	ImdbUrl    string  `gorm:"column:imdb_url;size:1024"`
	StoryLine  string  `gorm:"column:story_line;size:65535"`
	MoviePath  string  `gorm:"column:path;size:1024"`
	Tags       []Tag   `gorm:"-"`

	HasImage  bool    `gorm:"-"`
	Image     *[]byte `gorm:"-"`
	ImageId   int     `gorm:"column:image_id;"`
	ImagePath string  `gorm:"column:image_path;size:1024"`
	ToWatch   bool    `gorm:"column:to_watch"`
}

func (m *Movie) TableName() string {
	return "movies"
}

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

func (d *Database) GetAllMovies(searchFor string) ([]*Movie, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}

	var movies []*Movie

	var result *gorm.DB
	if searchFor == "" {
		if result = db.Order("title asc").Find(&movies); result.Error != nil {
			return nil, result.Error
		}
	} else {
		s := "%" + searchFor + "%"
		if result = db.Where("title like ? OR sub_title like ? OR year like ? OR story_line like ?", s, s, s, s).
			Order("title asc").
			Find(&movies); result.Error != nil {
			return nil, result.Error
		}
	}
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

func (d *Database) InsertMovie(movie *Movie) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}

	// Insert image
	if movie.HasImage && len(*movie.Image) > 0 {
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
}

func (d *Database) DeleteMovie(movie *Movie) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}
	if result := db.Delete(movie, movie.Id); result.Error != nil {
		return result.Error
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