package config

type GrpcConfig struct {
	ServerPort int `mapstructure:"server_port" yaml:"server_port"`
}
