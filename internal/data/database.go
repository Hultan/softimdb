package data

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Database struct {
	db              *gorm.DB
	cache           *ImageCache
	UseTestDatabase bool
}

func DatabaseNew(useTestDB bool) *Database {
	database := new(Database)
	database.UseTestDatabase = useTestDB
	database.cache = ImageCacheNew()
	return database
}

func (d *Database) getDatabase() (*gorm.DB, error) {
	if d.db == nil {
		db, err := d.OpenDatabase()
		if err != nil {
			return nil, err
		}
		d.db = db
	}
	return d.db, nil
}

func (d *Database) OpenDatabase() (*gorm.DB, error) {
	var connectionString = fmt.Sprintf("per:KnaskimGjwQ6M!@tcp(192.168.1.3:3306)/%s?parseTime=True", constDatabaseName)
	db, err := gorm.Open(mysql.New(mysql.Config{
		DriverName: "mysql",
		DSN:        connectionString,
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	//err = db.SetupJoinTable(&Movie{}, "Tags", &MovieTag{})
	//if err != nil {
	//	return nil, err
	//}
	return db, nil
}

func (d *Database) GetDatabaseName() string {
	if d.UseTestDatabase {
		return constDatabaseNameTest
	} else {
		return constDatabaseName
	}
}

func (d *Database) CloseDatabase() {
	if d.db == nil {
		return
	}
	d.db = nil

	return
}
