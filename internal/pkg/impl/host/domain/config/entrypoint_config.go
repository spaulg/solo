package config

type EntrypointConfig struct {
	HostEntrypointPath      string `mapstructure:"host_entrypoint_path" yaml:"host_entrypoint_path"`
	ContainerEntrypointPath string `mapstructure:"container_entrypoint_path" yaml:"container_entrypoint_path"`
}
