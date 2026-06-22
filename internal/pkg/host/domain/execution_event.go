package domain

type ExecutionEvent struct {
	CommandPath string   `json:"command_path"`
	CommandArgs []string `json:"command_args"`
	Error       string   `json:"error"`
	Complete    bool     `json:"complete"`
}

func NewExecutionEvent(commandPath string, commandArgs []string) *ExecutionEvent {
	return &ExecutionEvent{
		CommandPath: commandPath,
		CommandArgs: commandArgs,
		Complete:    false,
	}
}

func (t *ExecutionEvent) MarkComplete(res error) {
	if res != nil {
		t.Error = res.Error()
	}

	t.Complete = true
}
