package compose

type ToolConfig struct {
	Description      string  `mapstructure:"description" yaml:"description"`
	Command          string  `mapstructure:"command" yaml:"command"`
	Service          string  `mapstructure:"service" yaml:"service"`
	WorkingDirectory string  `mapstructure:"working_directory" yaml:"working_directory"`
	Shell            *string `yaml:"shell"`
}

type Tools map[string]ToolConfig
