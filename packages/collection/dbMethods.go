package collection

import (
	"fmt"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var db *gorm.DB

type gormIDer interface {
	gormID() uint
}

func initDb(dbFilepath string) {
	// Check if the file exists, and create it if it doesn't.
	if _, err := os.Stat(dbFilepath); os.IsNotExist(err) {
		file, err := os.Create(dbFilepath)
		if err != nil {
			panic(fmt.Sprintf("failed to create database file '%s'", dbFilepath))
		}
		file.Close()
	}

	var err error
	db, err = gorm.Open(sqlite.Open(dbFilepath), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database '%s'", dbFilepath))
	}
	db.AutoMigrate(&Collection{})
	db.AutoMigrate(&ColSelection{})
	db.AutoMigrate(&ProcessingEngineRunner{})
	db.AutoMigrate(&ProcessingEngine{})
	db.AutoMigrate(&Schema{})
}

type dbMethods struct{}

func (d *dbMethods) save(i interface{}) error {
	return db.Save(i).Error
}
func (d *dbMethods) delete(i interface{}) error {
	return db.Delete(i).Error
}

// i should be pointer to object (i = &object)
func (d *dbMethods) reload(i interface{}) error {
	ider := i.(gormIDer)
	id := ider.gormID()
	return db.Preload(clause.Associations).First(i, "id = ?", id).Error
}
