package cmd

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"vendingMaxine/packages/logger"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// appConfigType represents the config.yaml and env-vars
// Config file must exist and have all the fields set - missing fields will generate an error.
// Env-vars with prefix CONFIG_ will override config.yaml values (ex: export CONFIG_DEXCONFIG_SECRET="overriden from env-var")
// Sensitive values can be set in this file with a "fake" value, and then overriden with env-vars at runtime.
// See example with self-explanatory comments in file config.yaml

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

// validateYAMLSyntax checks if the YAML file is syntactically correct
func validateYAMLSyntax(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading YAML file: %w", err)
	}

	var temp interface{}
	if err := yaml.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("invalid YAML syntax: %w", err)
	}

	return nil
}

// validateConfig checks if any field in the config struct has a zero value
func validateConfig(cfg *appConfigType) error {
	v := reflect.ValueOf(cfg).Elem()

	var missingFields []string
	var validateStruct func(reflect.Value, string) error

	validateStruct = func(v reflect.Value, prefix string) error {
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fieldName := t.Field(i).Name
			fullPath := prefix + fieldName

			if field.Kind() == reflect.Struct {
				if err := validateStruct(field, fullPath+"."); err != nil {
					return err
				}
				continue
			}

			if field.IsZero() {
				missingFields = append(missingFields, fullPath)
			}
		}
		return nil
	}

	if err := validateStruct(v, ""); err != nil {
		return err
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("missing or empty configuration values for: %s", strings.Join(missingFields, ", "))
	}

	return nil
}

// loadAppConfig reads the configuration from a YAML file (config.yaml),
// and applies any environment variable overrides with the prefix "CONFIG_".
func loadAppConfig(configFilename string) (cfg *appConfigType, err error) {
	// Validate YAML syntax first
	if err := validateYAMLSyntax(configFilename); err != nil {
		log.Printf("❌ YAML validation: %v", err)
		return nil, fmt.Errorf("YAML validation failed: %w", err)
	}
	log.Printf("✓ YAML validation successful: %s", configFilename)

	// Use the config file specified by the flag
	viper.SetConfigFile(configFilename)
	viper.SetConfigType("yaml")

	// Set the environment variable prefix.
	viper.SetEnvPrefix("CONFIG")
	// Replace dots in keys with underscores to support nested configuration.
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	// Enable reading environment variables.
	viper.AutomaticEnv()

	// Read the config file.
	if err = viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file %s: %w", configFilename, err)
	}

	// Unmarshal the configuration into our Config struct.
	cfg = &appConfigType{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unable to decode configuration into struct: %w", err)
	}

	// Validate that all fields have non-zero values
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}
