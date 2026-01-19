package data

import (
	"fmt"
	"io"
	"os"
	"path"

	"gorm.io/gorm"
)

// image represents a movie image.
type image struct {
	Id   int    `gorm:"column:id;primary_key"`
	Data []byte `gorm:"column:image;"`
}

const imageCachePath = "/home/per/.cache/"

// TableName returns the name of the table.
func (i *image) TableName() string {
	return "image"
}

// createImage inserts an image into the database.
func (d *Database) createImage(image *image) error {
	db, err := d.getDatabase()
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}
	if err := db.Create(image).Error; err != nil {
		return fmt.Errorf("failed to insert image: %w", err)
	}

	return nil
}

// readImage returns an image from the database.
func (d *Database) readImage(imageId int) (*image, error) {
	img := image{}

	// Load from cache
	cachePath := path.Join(imageCachePath, "softimdb", fmt.Sprintf("%d.jpg", imageId))
	if d.existCachedImage(cachePath) {
		img.Id = imageId
		err := d.getCachedImage(&img, cachePath)
		if err == nil {
			return &img, nil
		}
	}

	db, err := d.getDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	if err := db.Where("Id=?", imageId).First(&img).Error; err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	// Store in cache
	if !d.existCachedImage(cachePath) {
		d.storeCachedImage(&img, cachePath)
	}

	return &img, nil
}

// UpdateImage replaces an image in the database.
func (d *Database) UpdateImage(movie *Movie, imageData []byte) error {
	db, err := d.getDatabase()
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	err = db.Transaction(
		func(tx *gorm.DB) error {
			if err := db.Delete(&image{}, movie.ImageId).Error; err != nil {
				return fmt.Errorf("failed to delete old image: %w", err)
			}

			img := &image{Data: imageData}
			if err := db.Create(img).Error; err != nil {
				return fmt.Errorf("failed to insert new image: %w", err)
			}

			if err := db.Model(&movie).Update("image_id", img.Id).Error; err != nil {
				return fmt.Errorf("failed to update image_id on movie: %w", err)
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

// deleteImage inserts an image into the database.
func (d *Database) deleteImage(movie *Movie) error {
	db, err := d.getDatabase()
	if err != nil {
		return fmt.Errorf("failed to get database: %w", err)
	}

	if err := db.Delete(&image{}, movie.ImageId).Error; err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	return nil
}

//
// Cached images
//

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
		return fmt.Errorf("failed to open cached image: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close failed: %w", closeErr)
		}
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to get cached image: %w", err)
	}
	image.Data = data
	return nil
}

func (d *Database) storeCachedImage(image *image, cachePath string) {
	if !d.existCachedImage(path.Join(imageCachePath, "softimdb")) {
		err := os.Mkdir(path.Join(imageCachePath, "softimdb"), os.ModePerm)
		if err != nil {
			return
		}
	}

	file, err := os.Create(cachePath)
	if err != nil {
		return
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close failed: %w", closeErr)
		}
	}()

	_, err = file.Write(image.Data)
	if err != nil {
		return
	}
}
