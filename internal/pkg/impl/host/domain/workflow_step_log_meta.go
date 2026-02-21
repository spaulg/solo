package domain

type WorkflowStepLogMeta struct {
	ExitCode         uint8    `json:"exit_code"`
	Command          string   `json:"command"`
	Arguments        []string `json:"arguments"`
	WorkingDirectory string   `json:"working_directory"`
}

func NewWorkflowStepLogMeta() *WorkflowStepLogMeta {
	return &WorkflowStepLogMeta{}
}

func (t *WorkflowStepLogMeta) SetExecutionInfo(exitCode uint8, command string, arguments []string, workingDirectory string) {
	t.ExitCode = exitCode
	t.Command = command
	t.Arguments = arguments
	t.WorkingDirectory = workingDirectory
}
