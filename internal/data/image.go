package data

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"gorm.io/gorm"
)

// image represents a movie image.
type image struct {
	Id   int    `gorm:"column:id;primary_key"`
	Data []byte `gorm:"column:image;"`
}

const imageCache = "/home/per/.cache/"

// TableName returns the name of the table.
func (i *image) TableName() string {
	return "image"
}

// UpdateImage replaces an image in the database.
func (d *Database) UpdateImage(movie *Movie, imageData []byte) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}

	err = db.Transaction(
		func(tx *gorm.DB) error {
			if result := db.Delete(&image{}, movie.ImageId); result.Error != nil {
				return result.Error
			}
			image := &image{Data: imageData}
			if result := db.Create(image); result.Error != nil {
				return result.Error
			}
			if result := db.Model(&movie).Update("image_id", image.Id); result.Error != nil {
				return result.Error
			}
			// TODO : Update cache

			return nil
		},
	)

	// Check transaction error
	if err != nil {
		return err
	}

	return nil
}

// getImage returns an image from the database.
func (d *Database) getImage(id int) (*image, error) {
	image := image{}

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

// deleteImage inserts an image into the database.
func (d *Database) deleteImage(movie *Movie) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}

	if result := db.Delete(&image{}, movie.ImageId); result.Error != nil {
		return result.Error
	}

	return nil
}

// insertImage inserts an image into the database.
func (d *Database) insertImage(image *image) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}
	if result := db.Create(image); result.Error != nil {
		return result.Error
	}

	return nil
}

//
// Cached images
//

func (d *Database) storeCachedImage(image *image, cachePath string) {
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
	defer func() {
		err = file.Close()
		if err != nil {
			log.Print(err)
		}
	}()

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
func (d *Database) getCachedImage(image *image, cachePath string) error {
	file, err := os.Open(cachePath)
	if err != nil {
		return err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			log.Print(err)
		}
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	image.Data = data
	return nil
}
