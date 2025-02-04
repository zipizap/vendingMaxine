package cmd

import (
	"vendingMaxine/packages/collectionPkg/dbCollectionPkg"
	"vendingMaxine/packages/gormCrud"
)

func dbInit() {
	// Initialize the database with a specific filename
	if err := gormCrud.InitializeDB("sqlite.db"); err != nil {
		panic(err)
	}

	// Migrate models
	err := gormCrud.MigrateModels(
		&dbCollectionPkg.DbCollection{},
	)
	if err != nil {
		panic(err)
	}
}

func Execute() {
	dbInit()
}
