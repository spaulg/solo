package domain

import (
	"github.com/spaulg/solo/internal/pkg/impl/host/domain/config"
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

type Config struct {
	StateDirectoryName string                     `mapstructure:"state_directory_name" yaml:"state_directory_name"`
	Logging            config.LoggingConfig       `mapstructure:"logging" yaml:"logging"`
	Entrypoint         config.EntrypointConfig    `mapstructure:"entrypoint" yaml:"entrypoint"`
	Orchestration      config.OrchestrationConfig `mapstructure:"orchestration" yaml:"orchestration"`
	Workflow           config.WorkflowConfig      `mapstructure:"workflow" yaml:"workflow"`
	Shell              config.ShellConfig         `mapstructure:"shell" yaml:"shell"`
}

func NewConfig() Config {
	return Config{
		StateDirectoryName: DefaultStateDirectoryName,

		Logging: config.LoggingConfig{
			Enabled: DefaultLoggingEnabled,
			Level:   DefaultLoggingLevel,
			Handler: DefaultLoggingHandler,
		},

		Entrypoint: config.EntrypointConfig{
			HostEntrypointPath:      DefaultHostEntrypoint,
			ContainerEntrypointPath: DefaultContainerEntrypoint,
		},

		Orchestration: config.OrchestrationConfig{
			SearchOrder: []string{DefaultDockerOrchestratorName},
			Orchestrators: map[string]config.OrchestratorConfig{
				DefaultDockerOrchestratorName: {
					Binary: DefaultDockerBinary,
				},
			},
		},

		Workflow: config.WorkflowConfig{
			Grpc: config.GrpcConfig{
				ServerPort: DefaultGrpcServerPort,
			},

			DefaultStepShell: DefaultStepShell,
		},

		Shell: config.ShellConfig{
			ShellPriority: DefaultShellPriority,
			DefaultShell:  DefaultShell,
		},
	}
}
