package config

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
	Entrypoint         Entrypoint    `mapstructure:"entrypoint"`
	Logging            LoggingConfig `mapstructure:"logging"`
	StateDirectoryName string        `mapstructure:"state_directory_name"`
	Orchestrator       string        `mapstructure:"orchestrator"`
	GrpcServerPort     int           `mapstructure:"grpc_server_port"`
}
