package dbPkg

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var Db *gorm.DB
var autoMigrateModels []interface{}

func AddModelForAutoMigration(model interface{}) {
	autoMigrateModels = append(autoMigrateModels, model)
}

func InitializeDb() (*gorm.DB, error) {
	var err error
	Db, err = gorm.Open(sqlite.Open("sqlite.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	for _, model := range autoMigrateModels {
		if err := Db.AutoMigrate(model); err != nil {
			return nil, err
		}
	}

	return Db, nil
}
