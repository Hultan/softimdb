package data

import (
	"fmt"
	"io"
	"os"
	"path"
)

// Image represents a movie image.
type Image struct {
	Id   int    `gorm:"column:id;primary_key"`
	Data []byte `gorm:"column:image;"`
}

const imageCache = "/home/per/.cache/"

// TableName returns the name of the table.
func (i *Image) TableName() string {
	return "image"
}

// GetImage returns an image from the database.
func (d *Database) GetImage(id int) (*Image, error) {
	image := Image{}

	// Load from cache
	cachePath := path.Join(imageCache, "softimdb", fmt.Sprintf("%d.png", id))
	if d.existCachedImage(cachePath) {
		image.Id = id
		err := d.getCachedImage(&image, cachePath)
		if err == nil {
			return &image, nil
		}
	}

	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}
	if result := db.Where("Id=?", id).First(&image); result.Error != nil {
		return nil, result.Error
	}

	// Store in cache
	if !d.existCachedImage(cachePath) {
		d.storeCachedImage(&image, cachePath)
	}

	return &image, nil
}

func (d *Database) storeCachedImage(image *Image, cachePath string) {
	if !d.existCachedImage(path.Join(imageCache, "softimdb")) {
		err := os.Mkdir(path.Join(imageCache, "softimdb"), os.ModePerm)
		if err != nil {
			return
		}
	}

	file, err := os.Create(cachePath)
	if err != nil {
		return
	}
	defer file.Close()

	_, err = file.Write(image.Data)
	if err != nil {
		return
	}
}

// existCachedImage returns true if an image exists in the cache.
func (d *Database) existCachedImage(cachePath string) bool {
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return false
	}
	return true
}

// getCachedImage returns an image from the cache.
func (d *Database) getCachedImage(image *Image, cachePath string) error {
	file, err := os.Open(cachePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	image.Data = data
	return nil
}

// InsertImage inserts an image into the database.
func (d *Database) InsertImage(image *Image) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}
	if result := db.Create(image); result.Error != nil {
		return result.Error
	}

	return nil
}

// UpdateImage updates an image.
func (d *Database) UpdateImage(image *Image) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}

	if result := db.Model(&image).Update("Data", image.Data); result.Error != nil {
		return result.Error
	}

	return nil
}
