package data

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

// Tag represents a movie tag.
type Tag struct {
	Id        int     `gorm:"column:id;primary_key"`
	Name      string  `gorm:"column:name;size:255"`
	IsPrivate bool    `gorm:"column:is_private;"`
	Movies    []Movie `gorm:"-"`
}

var tagCache *TagCache

// TableName returns the tag table name.
func (t *Tag) TableName() string {
	return "tag"
}

// GetTagByName returns a tag by name.
func (d *Database) GetTagByName(name string) (*Tag, error) {
	name = strings.Trim(name, " \t\n")

	// Check tag cache
	t := tagCache.GetByName(name)
	if t != nil {
		return t, nil
	}

	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}
	tag := Tag{}
	if result := db.Where("name=?", name).First(&tag); result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
		return nil, nil
	}

	return &tag, nil
}

// GetOrInsertTag either returns an existing tag or inserts a new tag and returns it.
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

	tag.Name = strings.Trim(tag.Name, " \t\n")

	// If it does not, create it
	if result := db.Create(tag); result.Error != nil {
		return nil, result.Error
	}
	return tag, nil
}

// GetTagsForMovie returns a list of tags connected to the given movie.
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
outerLoop:
	for i := range movieTags {
		// Check tag cache first
		t := tagCache.GetById(movieTags[i].TagId)
		if t != nil {
			tags = append(tags, *t)
			continue outerLoop
		}

		// Tag did not exist in the tag cache, load it
		// and add it to tag cache
		var tag Tag
		if result := db.Where("id=?", movieTags[i].TagId).Find(&tag); result.Error != nil {
			return nil, result.Error
		}
		tagCache.Add(&tag)
		tags = append(tags, tag)
	}

	return tags, nil
}

// GetTags returns all tags
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

	// Fill tag cache
	for i := range tags {
		tagCache.Add(&tags[i])
	}

	return tags, nil
}

// DeleteTagsForMovie deletes all tags for the given movie.
func (d *Database) DeleteTagsForMovie(movie *Movie) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}

	// Get tag id:s for movie
	var movieTags []MovieTag
	if result := db.Where("movie_id=?", movie.Id).Find(&movieTags); result.Error != nil {
		return result.Error
	}

	db.Delete(&movieTags)

	return nil
}
