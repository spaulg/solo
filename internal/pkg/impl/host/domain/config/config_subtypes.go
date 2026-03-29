package config

type LoggingConfig struct {
	Enabled bool   `mapstructure:"enabled" yaml:"enabled"`
	Level   string `mapstructure:"level" yaml:"level"`
	Handler string `mapstructure:"handler" yaml:"handler"`
}

type EntrypointConfig struct {
	HostEntrypointPath      string `mapstructure:"host_entrypoint_path" yaml:"host_entrypoint_path"`
	ContainerEntrypointPath string `mapstructure:"container_entrypoint_path" yaml:"container_entrypoint_path"`
}

type OrchestratorConfig struct {
	Binary string `mapstructure:"binary" yaml:"binary"`
}

type OrchestrationConfig struct {
	SearchOrder   []string                      `mapstructure:"search_order" yaml:"search_order"`
	Orchestrators map[string]OrchestratorConfig `mapstructure:"orchestrators" yaml:"orchestrators"`
}

type GrpcConfig struct {
	ServerPort int `mapstructure:"server_port" yaml:"server_port"`
}

type WorkflowConfig struct {
	Grpc             GrpcConfig `mapstructure:"grpc" yaml:"grpc"`
	DefaultStepShell string     `mapstructure:"default_step_shell" yaml:"default_step_shell"`
}

type ShellConfig struct {
	ShellPriority []string `mapstructure:"shell_priority" yaml:"shell_priority"`
	DefaultShell  string   `mapstructure:"default_shell" yaml:"default_shell"`
}
