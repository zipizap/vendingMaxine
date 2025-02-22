package cmd

import (
	"log"
	"vendingMaxine/packages/collectionPkg/dbCollectionPkg"
	"vendingMaxine/packages/gormCrud"
	"vendingMaxine/packages/webserver"
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

func appconfigInit() {
	cfg, err := loadAppConfig()
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
		&dbCollectionPkg.DbCollection{},
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
	appconfigInit()
	dbInit()
	webserverStart()
}
