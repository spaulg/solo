package wms

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type WorkflowEvent struct {
	eventFilePath string
	data          WorkflowEventData
}

type WorkflowEventData struct {
	CommandPath string   `json:"command_path"`
	CommandArgs []string `json:"command_args"`
	Error       string   `json:"error"`
	Complete    bool     `json:"complete"`
}

func NewWorkflowEvent(eventFilePath string, commandPath string, commandArgs []string) *WorkflowEvent {
	return &WorkflowEvent{
		eventFilePath: eventFilePath,
		data: WorkflowEventData{
			CommandPath: commandPath,
			CommandArgs: commandArgs,
			Complete:    false,
		},
	}
}

func (t *WorkflowEvent) MarkComplete(res error) {
	if res != nil {
		t.data.Error = res.Error()
	}

	t.data.Complete = true
}

func (t *WorkflowEvent) Persist() error {
	data, err := json.MarshalIndent(t.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal workflow event: %w", err)
	}

	if err := os.MkdirAll(path.Dir(t.eventFilePath), 0700); err != nil {
		return fmt.Errorf("failed to create workflow log directory: %w", err)
	}

	if err := os.WriteFile(t.eventFilePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write workflow event file: %w", err)
	}

	return nil
}
