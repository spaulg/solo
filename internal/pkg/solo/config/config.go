package config

import (
	"errors"
	"github.com/spf13/viper"
)

type LoggingConfig struct {
	Enabled bool
	Level   string
	Handler string
}

type Config struct {
	reader *viper.Viper

	Entrypoint     string
	LocalDirectory string
	Orchestrator   string
	GrpcServerPort uint16
	Logging        LoggingConfig
}

const (
	DefaultEntrypoint     = "/usr/local/bin/solo-entrypoint"
	DefaultLocalDirectory = "./.solo"
	DefaultOrchestrator   = "docker"
	DefaultGrpcServerPort = 0
	DefaultLoggingEnabled = false
	DefaultLoggingLevel   = "warning"
	DefaultLoggingHandler = "text"
)

func NewConfig() (*Config, error) {
	config := Config{
		Entrypoint:     DefaultEntrypoint,
		LocalDirectory: DefaultLocalDirectory,
		Orchestrator:   DefaultOrchestrator,
		GrpcServerPort: DefaultGrpcServerPort,

		Logging: LoggingConfig{
			Enabled: DefaultLoggingEnabled,
			Level:   DefaultLoggingLevel,
			Handler: DefaultLoggingHandler,
		},

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

	if err := config.unmarshallConfig(); err != nil {
		return nil, err
	}

	return &config, nil
}

func (config *Config) AddConfigPath(path string) error {
	config.reader.SetConfigName("solo-config")
	config.reader.SetConfigType("yaml")
	config.reader.AddConfigPath(path)

	if err := config.reader.MergeInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return err
		}
	}

	return config.unmarshallConfig()
}

func (config *Config) unmarshallConfig() error {
	if err := config.reader.Unmarshal(&config); err != nil {
		return err
	}

	return nil
}
