package config

type ConfigReader interface {
	AddConfigPath(path string) error
	GetConfig() *Config
}
