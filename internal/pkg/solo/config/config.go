package config

import (
	"errors"
	"github.com/spf13/viper"
)

type LoggingConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Level   string `mapstructure:"level"`
	Handler string `mapstructure:"handler"`
}

type Entrypoint struct {
	HostEntrypointPath      string `mapstructure:"host_entrypoint_path"`
	ContainerEntrypointPath string `mapstructure:"container_entrypoint_path"`
}

type Config struct {
	reader *viper.Viper

	Entrypoint         Entrypoint    `mapstructure:"entrypoint"`
	Logging            LoggingConfig `mapstructure:"logging"`
	StateDirectoryName string        `mapstructure:"state_directory_name"`
	Orchestrator       string        `mapstructure:"orchestrator"`
	GrpcServerPort     uint16        `mapstructure:"grpc_server_port"`
}

const (
	DefaultHostEntrypoint      = "/usr/local/bin/solo-entrypoint"
	DefaultContainerEntrypoint = "/usr/local/sbin/solo"
	DefaultStateDirectoryName  = "./.solo"
	DefaultOrchestrator        = "docker"
	DefaultGrpcServerPort      = 0
	DefaultLoggingEnabled      = true
	DefaultLoggingLevel        = "warning"
	DefaultLoggingHandler      = "text"
)

func NewConfig() (*Config, error) {
	config := Config{
		StateDirectoryName: DefaultStateDirectoryName,
		Orchestrator:       DefaultOrchestrator,
		GrpcServerPort:     DefaultGrpcServerPort,

		Entrypoint: Entrypoint{
			HostEntrypointPath:      DefaultHostEntrypoint,
			ContainerEntrypointPath: DefaultContainerEntrypoint,
		},

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
