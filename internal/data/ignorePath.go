package data

type IgnoredPath struct {
	Id               int    `gorm:"column:id;primary_key"`
	Path             string `gorm:"column:path;size:1024"`
	IgnoreCompletely bool   `gorm:"column:ignore_completely;"`
}

func (i *IgnoredPath) TableName() string {
	return "ignore_paths"
}

func (d *Database) GetAllIgnoredPaths() ([]*IgnoredPath, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}
	var ignoredPaths []*IgnoredPath
	if result := db.Find(&ignoredPaths); result.Error != nil {
		return nil, result.Error
	}

	return ignoredPaths, nil
}

func (d *Database) InsertIgnorePath(ignorePath *IgnoredPath) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}
	if result := db.Create(ignorePath); result.Error != nil {
		return result.Error
	}

	return nil
}

func (d *Database) DeleteIgnorePath(ignorePath *IgnoredPath) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}
	if result := db.Delete(ignorePath, ignorePath.Id); result.Error != nil {
		return result.Error
	}

	return nil
}
