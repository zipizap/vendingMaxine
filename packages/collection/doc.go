/*
collection package provides a model which manages Collection, ColSelection, ProcessingEngineRunner

Most usefull methods that are expected to be called from outside this package, are grouped as methods of Facilitator type.

Use the Facilitator functions, and avoid messing with anything else inside the package :)
(Facilitator is just a dummy type, a placeholder of exported functions)

	f, _ := collection.NewFacilitator()

	// First call f.InitSetup()
	dbFilepath := "./sqlite.db"
	catalogDir := "./catalog-default-0-0-0"
	zapSugaredLogger := myzapSuggaredLogger
	f.InitSetup(dbFilepath, catalogDir, zapSugaredLogger)


	// Then call any other Facilitator methods, ex:
	f.CollectionNew(...)
	f.CollectionsOverview(...)
	...
*/
package collection
