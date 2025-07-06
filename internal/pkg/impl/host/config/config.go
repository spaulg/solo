package config

import (
	config_types "github.com/spaulg/solo/internal/pkg/types/host/config"
)

const (
	DefaultHostEntrypoint         = "/usr/local/bin/solo-entrypoint"
	DefaultContainerEntrypoint    = "/usr/local/sbin/solo"
	DefaultStateDirectoryName     = "./.solo"
	DefaultGrpcServerPort         = 0
	DefaultLoggingEnabled         = true
	DefaultLoggingLevel           = "warning"
	DefaultLoggingHandler         = "text"
	DefaultDockerBinary           = "docker"
	DefaultDockerOrchestratorName = "docker"
)

func NewConfig() config_types.Config {
	return config_types.Config{
		StateDirectoryName: DefaultStateDirectoryName,
		GrpcServerPort:     DefaultGrpcServerPort,

		Entrypoint: config_types.Entrypoint{
			HostEntrypointPath:      DefaultHostEntrypoint,
			ContainerEntrypointPath: DefaultContainerEntrypoint,
		},

		Logging: config_types.LoggingConfig{
			Enabled: DefaultLoggingEnabled,
			Level:   DefaultLoggingLevel,
			Handler: DefaultLoggingHandler,
		},

		OrchestratorSearchOrder: []string{DefaultDockerOrchestratorName},
		Orchestrators: map[string]config_types.OrchestratorConfig{
			DefaultDockerOrchestratorName: {
				Binary: DefaultDockerBinary,
			},
		},
	}
}
