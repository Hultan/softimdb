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
	imageCache      *ImageCache
	UseTestDatabase bool
	config          *config.Config
	genreCache      *GenreCache
}

// DatabaseNew creates a new SoftIMDB Database object.
func DatabaseNew(useTestDB bool, config *config.Config) *Database {
	database := &Database{
		UseTestDatabase: useTestDB,
		config:          config,
		imageCache:      imageCacheNew(),
		genreCache:      genreCacheNew(),
	}

	return database
}

// CloseDatabase closes the underlying SQL database connection.
func (d *Database) CloseDatabase() {
	if d.db == nil {
		return
	}

	sqlDB, err := d.db.DB()
	if err != nil {
		log.Fatal("failed to get raw DB from GORM:", err)
	}

	if err := sqlDB.Close(); err != nil {
		log.Fatal("failed to close database connection:", err)
	}

	d.db = nil
}

func (d *Database) getDatabase() (*gorm.DB, error) {
	if d.db != nil {
		return d.db, nil
	}

	db, err := d.openDatabase()
	if err != nil {
		return nil, err
	}
	d.db = db

	return d.db, nil
}

func (d *Database) openDatabase() (*gorm.DB, error) {
	decryptedPassword, err := crypto.Decrypt(d.config.Database.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt password: %w", err)
	}

	// Create connection string
	var dsn = fmt.Sprintf(
		"%s:%s@tcp(%s:%v)/%s?parseTime=True",
		d.config.Database.User,
		decryptedPassword,
		d.config.Database.Server,
		d.config.Database.Port,
		d.config.Database.Database,
	)

	db, err := gorm.Open(mysql.New(mysql.Config{
		DriverName: "mysql",
		DSN:        dsn,
	}), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return db, nil
}
