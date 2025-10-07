package config

import (
	config_types "github.com/spaulg/solo/internal/pkg/types/host/config"
)

const (
	DefaultStateDirectoryName     = "./.solo"
	DefaultLoggingEnabled         = true
	DefaultLoggingLevel           = "warning"
	DefaultLoggingHandler         = "text"
	DefaultHostEntrypoint         = "/usr/local/bin/solo-entrypoint"
	DefaultContainerEntrypoint    = "/usr/local/sbin/solo"
	DefaultDockerBinary           = "docker"
	DefaultDockerOrchestratorName = "docker"
	DefaultGrpcServerPort         = 0
	DefaultShell                  = "/bin/sh"
	DefaultStepShell              = "/bin/sh"
)

// nolint:gochecknoglobals
var DefaultShellPriority = []string{
	"bash",
	"dash",
	"sh",
}

func NewConfig() config_types.Config {
	return config_types.Config{
		StateDirectoryName: DefaultStateDirectoryName,

		Logging: config_types.LoggingConfig{
			Enabled: DefaultLoggingEnabled,
			Level:   DefaultLoggingLevel,
			Handler: DefaultLoggingHandler,
		},

		Entrypoint: config_types.EntrypointConfig{
			HostEntrypointPath:      DefaultHostEntrypoint,
			ContainerEntrypointPath: DefaultContainerEntrypoint,
		},

		Orchestration: config_types.OrchestrationConfig{
			SearchOrder: []string{DefaultDockerOrchestratorName},
			Orchestrators: map[string]config_types.OrchestratorConfig{
				DefaultDockerOrchestratorName: {
					Binary: DefaultDockerBinary,
				},
			},
		},

		Workflow: config_types.WorkflowConfig{
			Grpc: config_types.GrpcConfig{
				ServerPort: DefaultGrpcServerPort,
			},

			DefaultStepShell: DefaultStepShell,
		},

		Shell: config_types.ShellConfig{
			ShellPriority: DefaultShellPriority,
			DefaultShell:  DefaultShell,
		},
	}
}
