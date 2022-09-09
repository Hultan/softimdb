package data

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/hultan/softteam/framework"
)

// Database represents a connection to the SoftIMDB database.
type Database struct {
	db              *gorm.DB
	cache           *ImageCache
	UseTestDatabase bool
}

// DatabaseNew creates a new SoftIMDB Database object.
func DatabaseNew(useTestDB bool) *Database {
	database := new(Database)
	database.UseTestDatabase = useTestDB
	database.cache = ImageCacheNew()
	return database
}

// CloseDatabase closes the database.
func (d *Database) CloseDatabase() {
	if d.db == nil {
		return
	}

	sqlDB, _ := d.db.DB()
	sqlDB.Close()

	d.db = nil

	return
}

func (d *Database) getPassword() string {
	fw := framework.NewFramework()
	passwordDecrypted, err := fw.Crypto.Decrypt(passwordEncrypted)
	if err != nil {
		return ""
	}
	return passwordDecrypted
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

	var connectionString = fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?parseTime=True",
		userName, d.getPassword(), serverIP, portNumber, constDatabaseName)

	db, err := gorm.Open(mysql.New(mysql.Config{
		DriverName: "mysql",
		DSN:        connectionString,
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
