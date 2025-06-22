package config

import (
	"errors"

	config_types "github.com/spaulg/solo/internal/pkg/types/host/config"
	"github.com/spf13/viper"
)

type ConfigReader struct {
	reader *viper.Viper
	config config_types.Config
}

func NewConfigReader() (config_types.ConfigReader, error) {
	config := ConfigReader{
		config: NewConfig(),
		reader: viper.New(),
	}

	config.reader.SetConfigName(".solo-config")
	config.reader.SetConfigType("yaml")
	config.reader.AddConfigPath("$HOME")

	if err := config.reader.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, err
		}
	}

	if err := config.unmarshalConfig(); err != nil {
		return nil, err
	}

	return &config, nil
}

func (t *ConfigReader) AddConfigPath(path string) error {
	t.reader.SetConfigName("solo-config")
	t.reader.SetConfigType("yaml")
	t.reader.AddConfigPath(path)

	if err := t.reader.MergeInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return err
		}
	}

	return t.unmarshalConfig()
}

func (t *ConfigReader) GetConfig() *config_types.Config {
	return &t.config
}

func (t *ConfigReader) unmarshalConfig() error {
	if err := t.reader.Unmarshal(&t.config); err != nil {
		return err
	}

	return nil
}
