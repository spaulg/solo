package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type WorkflowLogMeta struct {
	meta     map[string][]string
	filePath string
	file     *os.File
}

func LoadWorkflowMeta(workflowMetaPath string) (*WorkflowLogMeta, error) {
	workflowMetaFile, err := os.OpenFile(workflowMetaPath, os.O_RDWR|os.O_CREATE|os.O_SYNC, 0600)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to write step result to log file: failed to open workflow file %s: %w",
			workflowMetaPath,
			err,
		)
	}

	workflowMetaData, err := io.ReadAll(workflowMetaFile)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to write step result to log file: failed to read workflow file %s: %w",
			workflowMetaPath,
			err,
		)
	}

	workflowMeta := &WorkflowLogMeta{
		meta:     make(map[string][]string),
		filePath: workflowMetaPath,
		file:     workflowMetaFile,
	}

	if len(workflowMetaData) > 0 {
		if err := json.Unmarshal(workflowMetaData, &(workflowMeta.meta)); err != nil {
			return nil, fmt.Errorf(
				"failed to write step result to log file: failed to unmarshal json %w",
				err,
			)
		}
	}

	return workflowMeta, nil
}

func (t *WorkflowLogMeta) AppendStep(containerName string, stepID string) {
	t.meta[containerName] = append(t.meta[containerName], stepID)
}

func (t *WorkflowLogMeta) Persist() error {
	workflowMetaData, err := json.MarshalIndent(t.meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to write step result to log file: failed to marshal json: %w", err)
	}

	_, err = t.file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("failed to write step result to log file: failed to seek open meta file: %w", err)
	}

	if n, err := t.file.Write(workflowMetaData); err != nil || n != len(workflowMetaData) {
		return fmt.Errorf("failed to write workflow step meta output to file %s: %w", t.filePath, err)
	}

	return nil
}

func (t *WorkflowLogMeta) Close() error {
	return t.file.Close()
}
