/*
web package provides an http web+api server, which runs on top of package "vendingMaxine/packages/collection"

Most usefull methods that are expected to be called from outside this package, are grouped as methods of Facilitator type.

Use the Facilitator functions, and avoid messing with anything else inside the package :).
Facilitator is just a dummy type, a placeholder of functions :)

	f, _ := web.NewFacilitator()

	// First call f.InitSetup()
	dbFilepath := "./sqlite.db"
	processingEnginesDirpath := "./processingEngines"
	zapSugaredLogger := myzapSuggaredLogger
	f.InitSetup(dbFilepath, processingEnginesDirpath, zapSugaredLogger)

	//Then call any other Facilitator methods, ex:
	go f.StartServer(":8080")
	...
*/
package web
