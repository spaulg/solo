package config

type ShellConfig struct {
	ShellPriority []string `mapstructure:"shell_priority" yaml:"shell_priority"`
	DefaultShell  string   `mapstructure:"default_shell" yaml:"default_shell"`
}
