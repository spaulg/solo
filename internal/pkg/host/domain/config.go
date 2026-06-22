package domain

import (
	config2 "github.com/spaulg/solo/internal/pkg/host/domain/config"
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
	StateDirectoryName string                      `mapstructure:"state_directory_name" yaml:"state_directory_name"`
	Logging            config2.LoggingConfig       `mapstructure:"logging" yaml:"logging"`
	Entrypoint         config2.EntrypointConfig    `mapstructure:"entrypoint" yaml:"entrypoint"`
	Orchestration      config2.OrchestrationConfig `mapstructure:"orchestration" yaml:"orchestration"`
	Workflow           config2.WorkflowConfig      `mapstructure:"workflow" yaml:"workflow"`
	Shell              config2.ShellConfig         `mapstructure:"shell" yaml:"shell"`
}

func NewConfig() Config {
	return Config{
		StateDirectoryName: DefaultStateDirectoryName,

		Logging: config2.LoggingConfig{
			Enabled: DefaultLoggingEnabled,
			Level:   DefaultLoggingLevel,
			Handler: DefaultLoggingHandler,
		},

		Entrypoint: config2.EntrypointConfig{
			HostEntrypointPath:      DefaultHostEntrypoint,
			ContainerEntrypointPath: DefaultContainerEntrypoint,
		},

		Orchestration: config2.OrchestrationConfig{
			SearchOrder: []string{DefaultDockerOrchestratorName},
			Orchestrators: map[string]config2.OrchestratorConfig{
				DefaultDockerOrchestratorName: {
					Binary: DefaultDockerBinary,
				},
			},
		},

		Workflow: config2.WorkflowConfig{
			Grpc: config2.GrpcConfig{
				ServerPort: DefaultGrpcServerPort,
			},

			DefaultStepShell: DefaultStepShell,
		},

		Shell: config2.ShellConfig{
			ShellPriority: DefaultShellPriority,
			DefaultShell:  DefaultShell,
		},
	}
}
