package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type ExecutionEvent struct {
	eventFilePath string
	data          ExecutionEventData
}

type ExecutionEventData struct {
	CommandPath string   `json:"command_path"`
	CommandArgs []string `json:"command_args"`
	Error       string   `json:"error"`
	Complete    bool     `json:"complete"`
}

func NewExecutionEvent(eventFilePath string, commandPath string, commandArgs []string) *ExecutionEvent {
	return &ExecutionEvent{
		eventFilePath: eventFilePath,
		data: ExecutionEventData{
			CommandPath: commandPath,
			CommandArgs: commandArgs,
			Complete:    false,
		},
	}
}

func (t *ExecutionEvent) MarkComplete(res error) {
	if res != nil {
		t.data.Error = res.Error()
	}

	t.data.Complete = true
}

func (t *ExecutionEvent) Persist() error {
	data, err := json.MarshalIndent(t.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal execution event: %w", err)
	}

	if err := os.MkdirAll(path.Dir(t.eventFilePath), 0700); err != nil {
		return fmt.Errorf("failed to create execution log directory: %w", err)
	}

	if err := os.WriteFile(t.eventFilePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write execution event file: %w", err)
	}

	return nil
}
