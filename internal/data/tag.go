package data

import "gorm.io/gorm"

type Tag struct {
	Id        int     `gorm:"column:id;primary_key"`
	Name      string  `gorm:"column:name;size:255"`
	IsPrivate bool    `gorm:"column:is_private;"`
	Movies    []Movie `gorm:"-"`
}

type MovieTag struct {
	MovieId int `gorm:"column:movie_id;primary_key;"`
	TagId   int `gorm:"column:tag_id;primary_key;"`
}

func (t *Tag) TableName() string {
	return "tag"
}

func (m *MovieTag) TableName() string {
	return "movie_tag"
}

func (d *Database) GetTagByName(name string) (*Tag, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}
	tag := Tag{}
	if result := db.Where("name=?", name).First(&tag); result.Error != nil {
		if result.Error != gorm.ErrRecordNotFound {
			return nil, result.Error
		}
		return nil, nil
	}

	return &tag, nil
}

func (d *Database) GetOrInsertTag(tag *Tag) (*Tag, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}

	// Check if tag exists
	existingTag, err := d.GetTagByName(tag.Name)
	if err != nil {
		return nil, err
	}

	// If it does, return it
	if existingTag != nil {
		return existingTag, nil
	}

	// If it does not, create it
	if result := db.Create(tag); result.Error != nil {
		return nil, result.Error
	}
	return tag, nil
}

func (d *Database) InsertMovieTag(movie *Movie, tag *Tag) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}

	movieTag := MovieTag{MovieId: movie.Id, TagId: tag.Id}

	// If it does not, create it
	if result := db.Create(movieTag); result.Error != nil {
		return result.Error
	}
	return nil
}
func (d *Database) GetTagsForMovie(movie *Movie) ([]Tag, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}

	// Get tag id:s for movie
	var movieTags []MovieTag
	if result := db.Where("movie_id=?", movie.Id).Find(&movieTags); result.Error != nil {
		return nil, result.Error
	}

	var tags []Tag

	// Get tags for movieTags
	for i := range movieTags {
		// Get tag id:s for movie
		var tag Tag
		if result := db.Where("id=?", movieTags[i].TagId).Find(&tag); result.Error != nil {
			return nil, result.Error
		}
		tags = append(tags, tag)
	}

	return tags, nil
}


func (d *Database) GetTags() ([]Tag, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}

	// Get tag id:s for movie
	var tags []Tag
	if result := db.Find(&tags); result.Error != nil {
		return nil, result.Error
	}
	//
	//// Get tags for movieTags
	//for i := range movieTags {
	//	// Get tag id:s for movie
	//	var tag Tag
	//	if result := db.Where("id=?", movieTags[i].TagId).Find(&tag); result.Error != nil {
	//		return nil, result.Error
	//	}
	//	tags = append(tags, tag)
	//}

	return tags, nil
}
