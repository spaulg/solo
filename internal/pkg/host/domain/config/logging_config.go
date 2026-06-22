package config

type LoggingConfig struct {
	Enabled bool   `mapstructure:"enabled" yaml:"enabled"`
	Level   string `mapstructure:"level" yaml:"level"`
	Handler string `mapstructure:"handler" yaml:"handler"`
}
