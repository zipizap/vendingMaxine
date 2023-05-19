package config

import (
	"github.com/spf13/viper"
)

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
