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
	slog.Infof("Check if the dbFilepath '%s' exists, and create it if it doesn't", dbFilepath)
	if _, err := os.Stat(dbFilepath); os.IsNotExist(err) {
		file, err := os.Create(dbFilepath)
		if err != nil {
			slog.Fatalf(fmt.Sprintf("failed to create database file '%s'", dbFilepath))
		}
		file.Close()
	}

	var err error
	db, err = gorm.Open(sqlite.Open(dbFilepath), &gorm.Config{})
	if err != nil {
		slog.Fatalf("failed to connect database '%s'", dbFilepath)
	}
	db.AutoMigrate(&Blob{})
	db.AutoMigrate(&Catalog{})
	db.AutoMigrate(&Collection{})
	db.AutoMigrate(&ColSelection{})
	db.AutoMigrate(&ProcessingEngineRunner{})
}

type dbMethods struct{}

func (d *dbMethods) save(i interface{}) error {
	return db.Save(i).Error
}

// func (d *dbMethods) delete(i interface{}) error {
// 	return db.Delete(i).Error
// }

// i should be pointer to object (i = &object)
func (d *dbMethods) reload(i interface{}) error {
	/*
			GORM Nested Preloading
			Collection																0
			Collection.Catalog											        1
			Collection.ColSelections											1
			Collection.ColSelections.Collection										2
			Collection.ColSelections.Catalog										2
			Collection.ColSelections.ProcessingEngineRunner							2
			Collection.ColSelections.ProcessingEngineRunner.ColSelection		3

			Blob                                                                    0
	.	*/
	ider := i.(gormIDer)
	id := ider.gormID()

	switch i.(type) {
	case *Collection:
		return db.
			Preload(clause.Associations).        // direct-1-level-deep-fields are loaded
			Preload("ColSelections.Collection"). // 2orMore-level-deep-fields need explicit "nested preloading" for each deep association
			Preload("ColSelections.Catalog").
			Preload("ColSelections.ProcessingEngineRunner").
			Preload("ColSelections.ProcessingEngineRunner.ColSelection").
			First(i, "id = ?", id).Error
	case *ColSelection:
		return db.
			Preload(clause.Associations).                   // direct-1-level-deep-fields are loaded
			Preload("ProcessingEngineRunner.ColSelection"). // 2orMore-level-deep-fields need explicit "nested preloading" for each deep association
			First(i, "id = ?", id).Error

	default:
		// ProcessineEngineRunner, Catalog, Blob, ...
		// These dont have a 2-level-deep-field nested-association, so clause.Associations is enough to preload them
		return db.
			Preload(clause.Associations). // direct-1-level-deep-fields are loaded
			First(i, "id = ?", id).Error
	}
}
