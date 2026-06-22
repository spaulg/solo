package compose

type ToolConfig struct {
	DescriptionValue      string  `mapstructure:"description" yaml:"description"`
	CommandValue          string  `mapstructure:"command" yaml:"command"`
	ServiceValue          string  `mapstructure:"service" yaml:"service"`
	WorkingDirectoryValue string  `mapstructure:"working_directory" yaml:"working_directory"`
	ShellValue            *string `mapstructure:"shell" yaml:"shell"`
}

func (t ToolConfig) Description() string {
	return t.DescriptionValue
}

func (t ToolConfig) Command() string {
	return t.CommandValue
}

func (t ToolConfig) Service() string {
	return t.ServiceValue
}

func (t ToolConfig) WorkingDirectory() string {
	return t.WorkingDirectoryValue
}

func (t ToolConfig) Shell() *string {
	return t.ShellValue
}
