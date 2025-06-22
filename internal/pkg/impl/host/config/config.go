package config

import (
	config_types "github.com/spaulg/solo/internal/pkg/types/host/config"
)

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

func NewConfig() config_types.Config {
	return config_types.Config{
		StateDirectoryName: DefaultStateDirectoryName,
		Orchestrator:       DefaultOrchestrator,
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
	}
}
