package config

import (
	"github.com/spf13/viper"
)

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
func InitViperConfig(configFileBasename string, configFileDir string, EnvVarPrefix string, defaultValues map[string]string) error {
	viper.SetConfigName(configFileBasename)
	viper.AddConfigPath(configFileDir)
	viper.SetEnvPrefix(EnvVarPrefix)

	// Set default values
	for k, v := range defaultValues {
		viper.SetDefault(k, v)
	}

	// Enable reading from environment variables
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	return nil
}
