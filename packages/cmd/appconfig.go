package cmd

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

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
func loadAppConfig() (cfg *appConfigType, err error) {
	// Specify the config file name and type.
	viper.SetConfigName("config") // expects config.yaml
	viper.SetConfigType("yaml")
	// Add the current directory as the config file location.
	viper.AddConfigPath(".")

	// Set the environment variable prefix.
	viper.SetEnvPrefix("CONFIG")
	// Replace dots in keys with underscores to support nested configuration.
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	// Enable reading environment variables.
	viper.AutomaticEnv()

	// Read the config file.
	if err = viper.ReadInConfig(); err != nil {
		// If the config file is missing, decide whether to error out or proceed.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Optionally, you could set some defaults here if needed.
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
