package main

import (
	"DbObj/internal/collectionPkg"
	"DbObj/internal/dbPkg"

	"gorm.io/gorm"
)

var Db *gorm.DB

func init() {
	// set Db global var
	{
		var err error
		_, err = dbPkg.InitializeDb()
		if err != nil {
			panic("failed to connect database")
		}
	}

}

func main() {
	c, err := collectionPkg.CreateCollection("myCollectionName")
	if err != nil {
		panic(err)
	}
	println(c.Name())
}
