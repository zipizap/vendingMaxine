/*
# Config sources precedence:

	./config.yaml  >  env-vars  >  InitViperConfig()::defaultValues

# Example `./config.yaml`

		processingengines:
	  	  dirpath: "./processingEngines"

		db:
		  filepath: "./sqlite.db"

# Example env-vars

	VD_PROCESSINGENGINES_DIRPATH="./test/processingEngines"
	VD_DB_FILEPATH="./test/sqlite.db"

	Ex: VD_PROCESSINGENGINES_DIRPATH="./test/processingEngines" VD_DB_FILEPATH="./test/sqlite.db" go run main.go

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
package config
