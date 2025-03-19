package cmd

import (
	"vendingMaxine/packages/gormCrud"
	"vendingMaxine/packages/logger"
	"vendingMaxine/packages/models/dbModels"
	opshub "vendingMaxine/packages/opsHub"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var appConfig *appConfigType

func appconfigInit(flagConfigFilename *string) {
	cfg, err := loadAppConfig(*flagConfigFilename)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}
	appConfig = cfg
}

func loggerInit() {
	// Initialize logger with config
	logger.Init(&appConfig.Logging)
}

func showConfigInLogger() {
	log.Info().Interface("config", appConfig).Msg("Configuration loaded")
}

func dbInit() {
	// Initialize the database with a specific filename
	sqliteFilename := appConfig.Database.SqliteFilename
	log.Info().Str("filename", sqliteFilename).Msg("Initializing database")

	if err := gormCrud.InitializeDB(sqliteFilename); err != nil {
		log.Fatal().Err(err).Msg("failed to initialize database")
	}

	// Migrate models
	log.Debug().Msg("Migrating database models")
	err := gormCrud.MigrateModels(
		&dbModels.DbAccessPolicyMapping{},
		&dbModels.DbAccessPolicy{},
		&dbModels.DbRevState{},
		&dbModels.DbColRevision{},
		&dbModels.DbCollection{},
		// add here more models
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to migrate database models")
	}
	log.Info().Msg("Database migration completed")
}

func opsHubInit() {
	opshub.SetGlobalGroups(appConfig.GlobalGroups.GlobalAdminGroup, appConfig.GlobalGroups.GlobalReaderGroup)
}

func Execute() {
	var flagConfigFilename string
	rootCmd := &cobra.Command{
		Use:   "vendingMaxine",
		Short: "VendingMaxine application",
		Run: func(cmd *cobra.Command, args []string) {
			appconfigInit(&flagConfigFilename)
			loggerInit()
			showConfigInLogger()
			dbInit()
			opsHubInit()
			webserverStart()
		},
	}

	// Put additinoal command-line flags here
	rootCmd.PersistentFlags().StringVar(&flagConfigFilename, "config", "config.yaml", "config file path")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Error executing command")
	}
}
