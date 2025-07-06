package config

type LoggingConfig struct {
	Enabled bool   `mapstructure:"enabled" yaml:"enabled"`
	Level   string `mapstructure:"level" yaml:"level"`
	Handler string `mapstructure:"handler" yaml:"handler"`
}

type Entrypoint struct {
	HostEntrypointPath      string `mapstructure:"host_entrypoint_path" yaml:"host_entrypoint_path"`
	ContainerEntrypointPath string `mapstructure:"container_entrypoint_path" yaml:"container_entrypoint_path"`
}

type OrchestratorConfig struct {
	Binary string `mapstructure:"binary" yaml:"binary"`
}

type Config struct {
	Entrypoint              Entrypoint                    `mapstructure:"entrypoint" yaml:"entrypoint"`
	Logging                 LoggingConfig                 `mapstructure:"logging" yaml:"logging"`
	StateDirectoryName      string                        `mapstructure:"state_directory_name" yaml:"state_directory_name"`
	GrpcServerPort          int                           `mapstructure:"grpc_server_port" yaml:"grpc_server_port"`
	OrchestratorSearchOrder []string                      `mapstructure:"orchestrator_search_order" yaml:"orchestrator_search_order"`
	Orchestrators           map[string]OrchestratorConfig `mapstructure:"orchestrators" yaml:"orchestrators"`
}
