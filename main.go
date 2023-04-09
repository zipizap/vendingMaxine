package main

import (
	"vd-alpha/packages/collection"
	"vd-alpha/packages/config"

	"github.com/spf13/viper"
)

// . "vd-alpha/packages/xstate"

func init() {
	/*
			# PRECEDENCE: ./config.yaml  >  env-vars  >  InitViperConfig()::defaultValues
			# ./config.yaml

				processingengines:
				dirpath: "./processingEngines"

				db:
				filepath: "./sqlite.db"

			# env-vars

				VD_PROCESSINGENGINES_DIRPATH="./test/processingEngines" VD_DB_FILEPATH="./test/sqlite.db" go run main.go

			# How to read config-values in code:

		    // initialize
		    InitViperConfig(
				"config",
				".",
				"VD",
				map[string]string {
					"processingengines.dirpath": "./processingEngines",
					"db.filepath": "./sqlite.db",
				},
			)

			// in any moment, read config values:
			dbPeDirpath =: viper.GetString("processingengines.dirpath")
		    dbFilepath =: viper.GetString("db.filepath")

	*/
	config.InitViperConfig("config", ".", "VD",
		map[string]string{
			"processingengines.dirpath": "./processingEngines",
			"db.filepath":               "./sqlite.db",
		})
	processingEnginesDirpaths := viper.GetString("processingengines.dirpath")
	dbFilepath := viper.GetString("db.filepath")

	collection.InitSetup(dbFilepath, processingEnginesDirpaths)
}

func main() {
}
