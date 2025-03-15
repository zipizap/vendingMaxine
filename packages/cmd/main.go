package cmd

import (
	"vendingMaxine/packages/gormCrud"
	"vendingMaxine/packages/logger"
	"vendingMaxine/packages/models/dbModels"
	opshub "vendingMaxine/packages/opsHub"
	"vendingMaxine/packages/webserver"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// appConfigType represents the config.yaml and env-vars
// File config.yaml must exist and have all the fields set - missing fields will generate an error.
// Env-vars with prefix CONFIG_ are optional and can override config.yaml values
// Ex: export CONFIG_DEXCONFIG_SECRET="overriden from env-var"
/* Ex: config.yaml

Branding:
	Name: "My App"                    # String: Application display name
	Description: "My App Description" # String: Application description
	LogoPngFile: "myAppLogo.png"      # String: Path to logo PNG file

DexConfig:
	ClientId: example-app             # String: OAuth client ID
	ClientSecret: xxxxx               # String: OAuth client secret
	ClientRedirectURL: http://zzzzz   # String: Full URL for OAuth callbacks
	DexIssuer: http://yyyyyy          # String: Dex issuer URL

Database:
	SqliteFilename: "sqlite.db"       # String: Path to SQLite database file

Logging:
	Level: "debug"                    # String: "debug", "info", "warn", "error", "fatal", "panic"
	Type: "text-with-colors"          # String: "json", "text", or "text-with-colors"
	TimeFormat: "2006-01-02T15:04:05.999Z07:00"  # String: Go time format string
	Output: "console-and-file:/var/log/app.log"   # String: "console", "file:/path/to/file.log", or "console-and-file:/path/to/file.log"

GlobalGroups:
    GlobalAdminGroup: "0000-0000-1111" # String: Group ID for global admin group. Is not saved to db, loaded at startup
	GlobalReaderGroup: "0000-0000-2222" # String: Group ID for global reader group. Is not saved to db, loaded at startup
*/
// To improve this config, change this struct and nothing else
type appConfigType struct {
	Branding struct {
		Name        string `mapstructure:"Name"`
		Description string `mapstructure:"Description"`
		LogoPngFile string `mapstructure:"LogoPngFile"`
	} `mapstructure:"Branding"`
	DexConfig struct {
		ClientId          string `mapstructure:"ClientId"`
		ClientSecret      string `mapstructure:"ClientSecret"`
		ClientRedirectURL string `mapstructure:"ClientRedirectURL"`
		DexIssuer         string `mapstructure:"DexIssuer"`
	} `mapstructure:"DexConfig"`
	Database struct {
		SqliteFilename string `mapstructure:"SqliteFilename"`
	} `mapstructure:"Database"`
	Logging      logger.Config `mapstructure:"Logging"`
	GlobalGroups struct {
		GlobalAdminGroup  string `mapstructure:"GlobalAdminGroup"`
		GlobalReaderGroup string `mapstructure:"GlobalReaderGroup"`
	} `mapstructure:"GlobalGroups"`
}

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

func webserverStart() {
	log.Info().Msg("Starting web server")
	webserverConfigOauth := &webserver.ConfigOauthClientDex{
		ClientID:          appConfig.DexConfig.ClientId,
		ClientSecret:      appConfig.DexConfig.ClientSecret,
		ClientRedirectURL: appConfig.DexConfig.ClientRedirectURL,
		ClientScopes:      []string{"openid", "profile", "email", "groups"},
		DexIssuer:         appConfig.DexConfig.DexIssuer,
	}
	webserver.Start(webserverConfigOauth)
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
