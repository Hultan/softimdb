package data

// IgnoredPath represents the table IgnoredPath.
type IgnoredPath struct {
	Id               int    `gorm:"column:id;primary_key"`
	Path             string `gorm:"column:path;size:1024"`
	IgnoreCompletely bool   `gorm:"column:ignore_completely;"`
}

// TableName returns the table name.
func (i *IgnoredPath) TableName() string {
	return "ignore_paths"
}

// GetAllIgnoredPaths returns all ignored paths.
func (d *Database) GetAllIgnoredPaths() ([]*IgnoredPath, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}
	var ignoredPaths []*IgnoredPath
	if err := db.Find(&ignoredPaths).Error; err != nil {
		return nil, err
	}

	return ignoredPaths, nil
}

// InsertIgnorePath inserts a path to be ignored.
func (d *Database) InsertIgnorePath(ignorePath *IgnoredPath) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}
	if err := db.Create(ignorePath).Error; err != nil {
		return err
	}

	return nil
}

// DeleteIgnorePath deletes a path from the ignored paths
func (d *Database) DeleteIgnorePath(ignorePath *IgnoredPath) error {
	db, err := d.getDatabase()
	if err != nil {
		return err
	}
	if err := db.Delete(ignorePath, ignorePath.Id).Error; err != nil {
		return err
	}

	return nil
}
