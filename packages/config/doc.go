/*
# Config sources precedence:

	./config.yaml  >  env-vars  >  InitViperConfig()::defaultValues

	Ie, parameters in `./config.yaml` take precedence over env-vars. Env-vars take precedence over defaultValues.

## “./config.yaml`

			catalog:
	          default:
			    name: "default-0-0-0"
		  	    dirpath: "./catalog-default-0-0-0"

			db:
			  filepath: "./sqlite.db"

			logs:
			  # DEBUG or INFO (default) or ERROR
			  loglevel: "DEBUG"

## Env-vars

	VD_CATALOG_DEFAULT_NAME="default-0-0-0"
	VD_CATALOG_DEFAULT_DIRPATH="./catalog-default-0-0-0"
	VD_DB_FILEPATH="./test/sqlite.db"
	VD_LOGS_LOGLEVEL="DEBUG"

# How to read config-values in code:

	// initialize
	InitViperConfig(
		"config",
		".",
		"VD",
		map[string]string {
			"catalog.default.name": "the default value",
			"db.filepath": "./sqlite.db",
		},
	)

	// in any moment, read config values:
	dbPeDirpath =: viper.GetString("processingengines.dirpath")
	dbFilepath =: viper.GetString("db.filepath")
*/
package config
