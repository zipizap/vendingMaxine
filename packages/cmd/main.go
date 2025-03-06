package cmd

import (
	"log"
	"vendingMaxine/packages/gormCrud"
	"vendingMaxine/packages/models/dbModels"
	"vendingMaxine/packages/webserver"

	"github.com/spf13/cobra"
)

// appConfigType represents the config.yaml and env-vars
// File config.yaml must exist and have all the fields set - missing fields will generate an error.
// Env-vars with prefix CONFIG_ are optional and can override config.yaml values
// Ex: export CONFIG_DEXCONFIG_SECRET="overriden from env-var"
/* Ex: config.yaml

Branding:
  Name: "My App"
  LogoPngFile: "myAppLogo.png"

DexConfig:
  ClientId: example-app
  ClientSecret: xxxxx
  ClientRedirectURL: http://zzzzz
  DexIssuer: http://yyyyyy

Database:
  SqliteFilename: "sqlite.db"

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
}

var appConfig *appConfigType

func appconfigInit(flagConfigFilename *string) {
	cfg, err := loadAppConfig(*flagConfigFilename)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	appConfig = cfg
	// spew.Dump(cfg)
}

func dbInit() {
	// Initialize the database with a specific filename
	sqliteFilename := appConfig.Database.SqliteFilename
	if err := gormCrud.InitializeDB(sqliteFilename); err != nil {
		panic(err)
	}

	// Migrate models
	err := gormCrud.MigrateModels(
		&dbModels.DbCollection{},
		&dbModels.DbCollection{},
		&dbModels.DbAccessPolicy{},
		&dbModels.DbColRevision{},
		&dbModels.DbRevState{},
		// add here more models
	)
	if err != nil {
		panic(err)
	}
}

func webserverStart() {
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
			dbInit()
			webserverStart()
		},
	}

	// Put additinoal command-line flags here
	rootCmd.PersistentFlags().StringVar(&flagConfigFilename, "config", "config.yaml", "config file path")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}
