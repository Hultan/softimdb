package data

// Image represents a movie image.
type Image struct {
	Id   int    `gorm:"column:id;primary_key"`
	Data []byte `gorm:"column:image;"`
}

// TableName returns the name of the table.
func (i *Image) TableName() string {
	return "image"
}

// GetImage returns an image from the database.
func (d *Database) GetImage(id int) (*Image, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}
	image := Image{}
	if result := db.Where("Id=?", id).First(&image); result.Error != nil {
		return nil, result.Error
	}

	return &image, nil
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
