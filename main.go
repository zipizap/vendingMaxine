package main

import (
	"vendingMaxine/packages/collection"
	"vendingMaxine/packages/config"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var f *collection.Facilitator
var slog *zap.SugaredLogger

func Init_slog() error {
	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}
	defer logger.Sync() // flushes buffer, if any
	slog = logger.Sugar()
	slog.Info("Initialized logger")
	return nil
	/*
		slog.Infow("failed to fetch URL",
		  // Structured context as loosely typed key-value pairs.
		  "url", url,
		  "attempt", 3,
		  "backoff", time.Second,
		)
		slog.Infof("Failed to fetch URL: %s", url
	*/
}

func init() {
	err := Init_slog()
	if err != nil {
		panic(err)
	}

	slog.Info("Config - Initializing config from ./config.yaml or env-vars VD_PROCESSINGENGINES_DIRPATH,VD_DB_FILEPATH")
	config.InitViperConfig("config", ".", "VD",
		map[string]string{
			"processingengines.dirpath": "./processingEngines",
			"db.filepath":               "./sqlite.db",
		})
	processingEnginesDirpaths := viper.GetString("processingengines.dirpath")
	dbFilepath := viper.GetString("db.filepath")
	slog.Infof("Config - processingengines.dirpath: %s", processingEnginesDirpaths)
	slog.Infof("Config - db.filepath: %s", dbFilepath)

	slog.Info("Facilitator - creating new")
	f, err = collection.NewFacilitator()
	if err != nil {
		panic(err)
	}
	slog.Info("Facilitator - initializing")
	f.InitSetup(dbFilepath, processingEnginesDirpaths, slog)
}

func main() {
}
