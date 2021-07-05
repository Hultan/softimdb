package data

type Image struct {
	Id   int    `gorm:"column:id;primary_key"`
	Data *[]byte `gorm:"column:image;"`
}

func (i *Image) TableName() string {
	return "image"
}

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
