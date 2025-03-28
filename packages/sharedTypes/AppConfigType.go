package sharedTypes

import "vendingMaxine/packages/logger"

// appConfigType represents the config.yaml and env-vars
// Config file must exist and have all the fields set - missing fields will generate an error.
// Env-vars with prefix CONFIG_ will override config.yaml values (ex: export CONFIG_DEXCONFIG_SECRET="overriden from env-var")
// Sensitive values can be set in this file with a "fake" value, and then overriden with env-vars at runtime.
// See example with self-explanatory comments in file config.yaml
//
// To improve this config, change this struct and nothing else
type AppConfigType struct {
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
