package compose

type WorkflowStep struct {
	NameValue             string  `mapstructure:"name" yaml:"name"`
	RunValue              string  `mapstructure:"run" yaml:"run"`
	ShellValue            *string `mapstructure:"shell" yaml:"shell"`
	WorkingDirectoryValue *string `mapstructure:"working_dir" yaml:"working_dir"`
}

func NewWorkflowStep(name string, run string, shell *string, workingDirectory *string) WorkflowStep {
	return WorkflowStep{
		NameValue:             name,
		RunValue:              run,
		ShellValue:            shell,
		WorkingDirectoryValue: workingDirectory,
	}
}

func (t WorkflowStep) Name() string {
	return t.NameValue
}

func (t WorkflowStep) Run() string {
	return t.RunValue
}

func (t WorkflowStep) Shell() *string {
	return t.ShellValue
}

func (t WorkflowStep) WorkingDirectory() *string {
	return t.WorkingDirectoryValue
}
