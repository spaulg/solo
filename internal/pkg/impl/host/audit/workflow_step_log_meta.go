package audit

import (
	"encoding/json"
	"fmt"
	"os"
)

type WorkflowStepLogMeta struct {
	ExitCode         uint8    `json:"exit_code"`
	Command          string   `json:"command"`
	Arguments        []string `json:"arguments"`
	WorkingDirectory string   `json:"working_directory"`
}

func NewStepLogMeta(exitCode uint8, command string, arguments []string, workingDirectory string) WorkflowStepLogMeta {
	return WorkflowStepLogMeta{
		ExitCode:         exitCode,
		Command:          command,
		Arguments:        arguments,
		WorkingDirectory: workingDirectory,
	}
}

func (t *WorkflowStepLogMeta) Persist(metaPath string) error {
	metaData, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to write step result to log file: failed to marshal json file: %w", err)
	}

	metaFile, err := os.OpenFile(metaPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("failed to write step result to log file: failed to open metadata file %s: %w", metaPath, err)
	}

	defer func(metaFile *os.File) {
		_ = metaFile.Close()
	}(metaFile)

	if n, err := metaFile.Write(metaData); err != nil || n != len(metaData) {
		return fmt.Errorf("failed to write complete step meta output to file %s: %w", metaPath, err)
	}

	return nil
}
