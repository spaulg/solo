package config

type WorkflowConfig struct {
	Grpc             GrpcConfig `mapstructure:"grpc" yaml:"grpc"`
	DefaultStepShell string     `mapstructure:"default_step_shell" yaml:"default_step_shell"`
}
