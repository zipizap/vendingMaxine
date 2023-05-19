/*
How to use this package in your code:

	import .../collection

	// Most usefull methods that are expected to be called by an API, are grouped as methods of Facilitator type
	// Use the Facilitator functions, and avoid messing with anything else inside the package :)
	// (Facilitator is just a dummy type, a placeholder of functions to be called by the API :))
	f, _ := collection.NewFacilitator()

	// First call f.InitSetup()
	dbFilepath := "./sqlite.db"
	processingEnginesDirpath := "./processingEngines"
	zapSugaredLogger := myzapSuggaredLogger
	f.InitSetup(dbFilepath, processingEnginesDirpath, zapSugaredLogger)


	// Then call any other Facilitator methods, ex:
	f.CollectionNew(...)
	f.CollectionsOverview(...)
	...
*/
package collection
