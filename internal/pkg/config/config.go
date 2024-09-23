package config

import (
	"errors"
	"github.com/spaulg/solo/internal/pkg/project_file"
	"github.com/spf13/viper"
)

func ReadConfig(projectFile *project_file.ProjectFile) (*viper.Viper, error) {
	config := viper.New()

	config.SetDefault("Entrypoint", "../prototype/solo-entrypoint.sh")
	config.SetDefault("LocalDirectory", ".solo")

	config.SetConfigName("config")
	config.SetConfigType("yaml")
	config.AddConfigPath("$HOME/.solo")

	if projectFile != nil {
		config.AddConfigPath(projectFile.Directory)
	}

	if err := config.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, err
		}
	}

	return config, nil
}
