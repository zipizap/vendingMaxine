package main

import (
	"DbObj/internal/collectionPkg"
	"DbObj/internal/dbPkg"
	"fmt"
)

func init() {
	// init() of main package is called after all the init() of imported packages

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
	myCollectionName := "myCollectionName"
	var col *collectionPkg.Collection

	fmt.Println("Creating a new collection:", myCollectionName)
	{
		var err error
		col, err = collectionPkg.CreateCollection(myCollectionName)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("Reading the collection name")
	{
		readName, err := col.GetName()
		if err != nil {
			panic(err)
		}
		fmt.Println("Collection name is:", readName)
	}

	fmt.Println("Setting the collection name to 'newCollectionName'")
	myCollectionName = "newCollectionName"
	{
		err := col.SetName(myCollectionName)
		if err != nil {
			panic(err)
		}
		readName, err := col.GetName()
		if err != nil {
			panic(err)
		}
		fmt.Println("Collection name is:", readName)
	}

	var col2 *collectionPkg.Collection
	fmt.Println("Loading collection by name:", myCollectionName)
	{
		var err error
		col2, err = collectionPkg.LoadCollectionByName(myCollectionName)
		if err != nil {
			panic(err)
		}
		readName, err := col2.GetName()
		if err != nil {
			panic(err)
		}
		fmt.Println("Collection name is:", readName)
	}

	fmt.Println("Setting the collection name to 'newCollectionName2'")
	myCollectionName = "newCollectionName2"
	{
		err := col2.SetName(myCollectionName)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("Reading the collection name from col")
	{
		readName, err := col.GetName()
		if err != nil {
			panic(err)
		}
		fmt.Println("col - Collection name is:", readName)
	}

}
