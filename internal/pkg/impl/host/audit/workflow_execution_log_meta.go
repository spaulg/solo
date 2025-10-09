package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	wms_types "github.com/spaulg/solo/internal/pkg/types/host/wms/logging"
)

type WorkflowExecutionLogMeta struct {
	eventFilePath string

	CommandPath string   `json:"command_path"`
	CommandArgs []string `json:"command_args"`
	Error       string   `json:"error"`
	Complete    bool     `json:"complete"`
}

func NewWorkflowEvent(eventFilePath string, commandPath string, commandArgs []string) wms_types.WorkflowExecutionLogMeta {
	return &WorkflowExecutionLogMeta{
		eventFilePath: eventFilePath,
		CommandPath:   commandPath,
		CommandArgs:   commandArgs,
		Complete:      false,
	}
}

func LoadWorkflowEvent(eventFilePath string) (*WorkflowExecutionLogMeta, error) {
	eventData, err := os.ReadFile(eventFilePath)
	if err != nil {
		return nil, err
	}

	event := WorkflowExecutionLogMeta{}
	err = json.Unmarshal(eventData, &event)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func (t *WorkflowExecutionLogMeta) MarkComplete(res error) {
	if res != nil {
		t.Error = res.Error()
	}

	t.Complete = true
}

func (t *WorkflowExecutionLogMeta) Persist() error {
	data, err := json.MarshalIndent(t, "", "  ")
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

func (t *WorkflowExecutionLogMeta) GetCommandPath() string {
	return t.CommandPath
}

func (t *WorkflowExecutionLogMeta) GetCommandArgs() []string {
	return t.CommandArgs
}

func (t *WorkflowExecutionLogMeta) GetError() string {
	return t.Error
}

func (t *WorkflowExecutionLogMeta) GetComplete() bool {
	return t.Complete
}
