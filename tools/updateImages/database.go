package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/hultan/crypto"
	"github.com/hultan/softimdb/internal/config"
)

// Database represents a connection to the SoftIMDB database.
type Database struct {
	db     *gorm.DB
	config *config.Config
}

// DatabaseNew creates a new SoftIMDB Database object.
func DatabaseNew(config *config.Config) *Database {
	database := &Database{
		config: config,
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
		if err := d.isOpen(); err != nil {
			return nil, err
		}

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

func (d *Database) isOpen() error {
	sqlDB, err := d.db.DB() // Get the underlying *sql.DB
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

func (d *Database) getQuery(sqlJoin string, sqlWhere string, sqlArgs map[string]interface{}, sqlOrderBy string) (*gorm.DB, error) {
	db, err := d.getDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	if sqlJoin != "" {
		db = db.Joins(sqlJoin)
	}
	if sqlWhere != "" {
		if len(sqlArgs) == 0 {
			db = db.Where(sqlWhere)
		} else {
			db = db.Where(sqlWhere, sqlArgs)
		}
	}

	db = db.Order(sqlOrderBy)

	return db, nil
}

// saveImage returns an image from the database.
func (d *Database) saveImage(id int) (*Poster, error) {
	image := Poster{}

	db, err := d.getDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	if err := db.Where("Id=?", id).First(&image).Error; err != nil {
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	dest := fmt.Sprintf("/home/per/temp/softimdb/fixed/%d.jpg", id)
	if err := os.WriteFile(dest, image.Data, 0o644); err != nil {
		log.Fatalf("cannot write %s: %v", dest, err)
	}
	log.Printf("Saved %d bytes to %s", len(image.Data), dest)
	return &image, nil
}
