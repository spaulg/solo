package audit

type StepLogMeta struct {
	ExitCode         uint8    `json:"exit_code"`
	Command          string   `json:"command"`
	Arguments        []string `json:"arguments"`
	WorkingDirectory string   `json:"working_directory"`
}
