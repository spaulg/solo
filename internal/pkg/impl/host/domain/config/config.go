package config

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
	StateDirectoryName string              `mapstructure:"state_directory_name" yaml:"state_directory_name"`
	Logging            LoggingConfig       `mapstructure:"logging" yaml:"logging"`
	Entrypoint         EntrypointConfig    `mapstructure:"entrypoint" yaml:"entrypoint"`
	Orchestration      OrchestrationConfig `mapstructure:"orchestration" yaml:"orchestration"`
	Workflow           WorkflowConfig      `mapstructure:"workflow" yaml:"workflow"`
	Shell              ShellConfig         `mapstructure:"shell" yaml:"shell"`
}

func NewConfig() Config {
	return Config{
		StateDirectoryName: DefaultStateDirectoryName,

		Logging: LoggingConfig{
			Enabled: DefaultLoggingEnabled,
			Level:   DefaultLoggingLevel,
			Handler: DefaultLoggingHandler,
		},

		Entrypoint: EntrypointConfig{
			HostEntrypointPath:      DefaultHostEntrypoint,
			ContainerEntrypointPath: DefaultContainerEntrypoint,
		},

		Orchestration: OrchestrationConfig{
			SearchOrder: []string{DefaultDockerOrchestratorName},
			Orchestrators: map[string]OrchestratorConfig{
				DefaultDockerOrchestratorName: {
					Binary: DefaultDockerBinary,
				},
			},
		},

		Workflow: WorkflowConfig{
			Grpc: GrpcConfig{
				ServerPort: DefaultGrpcServerPort,
			},

			DefaultStepShell: DefaultStepShell,
		},

		Shell: ShellConfig{
			ShellPriority: DefaultShellPriority,
			DefaultShell:  DefaultShell,
		},
	}
}
