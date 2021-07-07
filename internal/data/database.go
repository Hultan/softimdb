package data

import (
	"bufio"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
)

type Database struct {
	db              *gorm.DB
	cache           *ImageCache
	UseTestDatabase bool
}

const (
	userName        = "per"
	credentialsFile = "/home/per/.config/softteam/softimdb/.credentials"
	serverIP        = "192.168.1.3"
	portNumber      = 3306
)

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

	var connectionString = fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?parseTime=True",
		userName, d.getPassword(), serverIP, portNumber, constDatabaseName)

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

func (d *Database) getPassword() string {
	file, err := os.Open(credentialsFile)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	// Create a new scanner
	scanner := bufio.NewScanner(file)
	// Split by lines
	scanner.Split(bufio.ScanLines)
	// Scan the file
	scanner.Scan()
	// Read the first line (this is a single line file)
	password := scanner.Text()

	return password
}
