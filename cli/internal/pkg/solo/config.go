package solo

import (
	"errors"
	"github.com/spf13/viper"
)

func NewConfig(projectFile *Project) (*Config, error) {
	configReader := viper.New()

	configReader.SetConfigName(".solo-config")
	configReader.SetConfigType("yaml")
	configReader.AddConfigPath("$HOME")

	if err := configReader.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, err
		}
	}

	if projectFile != nil {
		configReader.SetConfigName("solo-config")
		configReader.SetConfigType("yaml")
		configReader.AddConfigPath(projectFile.Directory)

		if err := configReader.MergeInConfig(); err != nil {
			var configFileNotFoundError viper.ConfigFileNotFoundError
			if !errors.As(err, &configFileNotFoundError) {
				return nil, err
			}
		}
	}

	config := Config{
		Entrypoint:     "/usr/lib/solo/solo-entrypoint",
		LocalDirectory: "./.solo",
	}

	if err := configReader.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
