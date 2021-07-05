package data

type Setting struct {
	Id       int    `gorm:"column:id;primary_key"`
	BasePath string `gorm:"column:base_path;size:1024"`
	IsSamba  bool   `gorm:"column:is_samba;"`
}

func (s *Setting) TableName() string {
	return "settings"
}

func (d *Database) GetSettings() (*Setting, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, err
	}
	setting := Setting{}
	if result := db.First(&setting); result.Error != nil {
		return nil, result.Error
	}

	return &setting, nil
}
