package data

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/hultan/crypto"
	"github.com/hultan/softimdb/internal/config"
)

// Database represents a connection to the SoftIMDB database.
type Database struct {
	db              *gorm.DB
	cache           *ImageCache
	UseTestDatabase bool
	config          *config.Config
}

// DatabaseNew creates a new SoftIMDB Database object.
func DatabaseNew(useTestDB bool, config *config.Config) *Database {
	database := &Database{}
	database.UseTestDatabase = useTestDB
	database.cache = imageCacheNew()
	database.config = config

	tagCache = NewTagCache()

	return database
}

// CloseDatabase closes the database.
func (d *Database) CloseDatabase() {
	if d.db == nil {
		return
	}

	sqlDB, _ := d.db.DB()
	err := sqlDB.Close()
	if err != nil {
		log.Fatal(err)
	}

	d.db = nil

	return
}

func (d *Database) getDatabase() (*gorm.DB, error) {
	// d.CloseDatabase()

	if d.db == nil {
		db, err := d.openDatabase()
		if err != nil {
			return nil, err
		}
		d.db = db
	}
	return d.db, nil
}

func (d *Database) openDatabase() (*gorm.DB, error) {
	passwordDecrypted, err := crypto.Decrypt(d.config.Database.Password)
	if err != nil {
		return nil, err
	}

	var connectionString = fmt.Sprintf(
		"%s:%s@tcp(%s:%v)/%s?parseTime=True",
		d.config.Database.User,
		passwordDecrypted,
		d.config.Database.Server,
		d.config.Database.Port,
		d.config.Database.Database,
	)

	db, err := gorm.Open(
		mysql.New(
			mysql.Config{
				DriverName: "mysql",
				DSN:        connectionString,
			},
		), &gorm.Config{},
	)
	if err != nil {
		return nil, err
	}
	return db, nil
}
