package audit

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"sync"

	"github.com/spaulg/solo/internal/pkg/impl/host/context"
	audit_types "github.com/spaulg/solo/internal/pkg/types/host/audit"
	events_types "github.com/spaulg/solo/internal/pkg/types/host/events"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/wms"
)

const auditLogsPath = "audit_logs"
const executionEventFile = "event.json"

type Auditor struct {
	soloCtx         *context.CliContext
	mu              sync.RWMutex
	outputDirectory string
}

func NewAuditor(soloCtx *context.CliContext) audit_types.Auditor {
	outputDirectory := path.Join(
		soloCtx.Project.GetStateDirectoryRoot(),
		auditLogsPath,
		soloCtx.TriggerDateTime.Format("2006-01-02T15-04-05.999999999Z"),
	)

	return &Auditor{
		soloCtx:         soloCtx,
		outputDirectory: outputDirectory,
	}
}

func (t *Auditor) RecordExecutionEvent(callback func() error) error {
	eventFile := path.Join(t.outputDirectory, executionEventFile)
	workflowEvent := NewExecutionEvent(eventFile, t.soloCtx.CommandPath, t.soloCtx.CommandArgs)
	if err := workflowEvent.Persist(); err != nil {
		return fmt.Errorf("failed to persist workflow event: %w", err)
	}

	res := callback()
	workflowEvent.MarkComplete(res)

	if err := workflowEvent.Persist(); err != nil {
		return fmt.Errorf("failed to record workflow event complete: %w", err)
	}

	return res
}

func (t *Auditor) Publish(event events_types.Event) {
	switch e := event.(type) {
	case *wms_types.WorkflowStepOutputEvent:
		t.writeStepOutput(e)

	case *wms_types.WorkflowStepCompleteEvent:
		t.writeStepResult(e)
	}
}

func (t *Auditor) writeStepOutput(e *wms_types.WorkflowStepOutputEvent) {
	if e.Stderr == "" && e.Stdout == "" {
		return
	}

	outputDirectory := path.Join(
		t.outputDirectory,
		e.WorkflowName.String(),
	)

	_, err := os.Stat(outputDirectory)
	if errors.Is(err, fs.ErrNotExist) {
		if err := os.MkdirAll(outputDirectory, 0700); err != nil {
			t.soloCtx.Logger.Error(fmt.Sprintf(
				"Failed to write step output to log files: failed to create directory %s: %v",
				outputDirectory,
				err,
			))

			return
		}
	}

	if e.Stderr != "" {
		stderrPath := path.Join(outputDirectory, e.StepID+".stderr")
		t.appendStepOutputFile(stderrPath, e.Stderr)

		combinedOutputPath := path.Join(outputDirectory, e.StepID+".out")
		t.appendStepOutputFile(combinedOutputPath, e.Stderr)
	}

	if e.Stdout != "" {
		stdoutPath := path.Join(outputDirectory, e.StepID+".stdout")
		t.appendStepOutputFile(stdoutPath, e.Stdout)

		combinedOutputPath := path.Join(outputDirectory, e.StepID+".out")
		t.appendStepOutputFile(combinedOutputPath, e.Stdout)
	}
}

func (t *Auditor) appendStepOutputFile(
	outputFilePath string,
	output string,
) {
	outputFile, err := os.OpenFile(outputFilePath, os.O_SYNC|os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		t.soloCtx.Logger.Error(fmt.Sprintf(
			"Failed to write step output to log file: failed to open output file: %s: %v",
			outputFilePath,
			err,
		))

		return
	}

	defer func(outputFile *os.File) {
		if err = outputFile.Close(); err != nil {
			t.soloCtx.Logger.Error(fmt.Sprintf(
				"Failed to close step output file: %v",
				err,
			))
		}
	}(outputFile)

	if n, err := outputFile.WriteString(output); err != nil || n != len(output) {
		t.soloCtx.Logger.Error(fmt.Sprintf(
			"Failed to write complete step output to file: %s: %v",
			outputFilePath,
			err,
		))
	}
}

func (t *Auditor) writeStepResult(e *wms_types.WorkflowStepCompleteEvent) {
	outputDirectory := path.Join(
		t.outputDirectory,
		e.WorkflowName.String(),
	)

	_, err := os.Stat(outputDirectory)
	if errors.Is(err, fs.ErrNotExist) {
		if err := os.MkdirAll(outputDirectory, 0700); err != nil {
			t.soloCtx.Logger.Error(fmt.Sprintf(
				"Failed to write step result to log files: failed to create directory %s: %v",
				outputDirectory,
				err,
			))

			return
		}
	}

	metaPath := path.Join(outputDirectory, e.StepID+".meta.json")
	metaJSON := StepLogMeta{
		ExitCode:         e.ExitCode,
		Command:          e.Command,
		Arguments:        e.Arguments,
		WorkingDirectory: e.Cwd,
	}

	metaData, err := json.MarshalIndent(metaJSON, "", "  ")
	if err != nil {
		t.soloCtx.Logger.Error(fmt.Sprintf(
			"Failed to write step result to log file: failed to marshal json file: %v",
			err,
		))

		return
	}

	metaFile, err := os.OpenFile(metaPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		t.soloCtx.Logger.Error(fmt.Sprintf(
			"Failed to write step result to log file: failed to open metadata file: %s: %v",
			metaPath,
			err,
		))

		return
	}

	defer func(metaFile *os.File) {
		if err = metaFile.Close(); err != nil {
			t.soloCtx.Logger.Error(fmt.Sprintf(
				"Failed to close step meta file: %v",
				err,
			))
		}
	}(metaFile)

	if n, err := metaFile.Write(metaData); err != nil || n != len(metaData) {
		t.soloCtx.Logger.Error(fmt.Sprintf(
			"Failed to write complete step meta output to file: %s: %v",
			metaPath,
			err,
		))
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	workflowMetaPath := path.Join(outputDirectory, e.WorkflowName.String()+".meta.json")

	workflowMetaFile, err := os.OpenFile(workflowMetaPath, os.O_RDWR|os.O_CREATE|os.O_SYNC, 0600)
	if err != nil {
		t.soloCtx.Logger.Error(fmt.Sprintf(
			"Failed to write step result to log file: failed to open workflow file: %s: %v",
			workflowMetaPath,
			err,
		))

		return
	}

	defer func(workflowMetaFile *os.File) {
		if err = workflowMetaFile.Close(); err != nil {
			t.soloCtx.Logger.Error(fmt.Sprintf(
				"Failed to close workflow meta file: %v",
				err,
			))
		}
	}(workflowMetaFile)

	workflowMetaData, err := io.ReadAll(workflowMetaFile)
	if err != nil {
		t.soloCtx.Logger.Error(fmt.Sprintf(
			"Failed to write step result to log file: failed to read workflow file: %s: %v",
			workflowMetaPath,
			err,
		))

		return
	}

	workflowMeta := make(WorkflowMeta)

	if len(workflowMetaData) > 0 {
		if err := json.Unmarshal(workflowMetaData, &workflowMeta); err != nil {
			t.soloCtx.Logger.Error(fmt.Sprintf(
				"Failed to write step result to log file: failed to unmarshal json: %v",
				err,
			))

			return
		}
	}

	workflowMeta[e.ContainerName] = append(workflowMeta[e.ContainerName], e.StepID)

	workflowMetaData, err = json.MarshalIndent(workflowMeta, "", "  ")
	if err != nil {
		t.soloCtx.Logger.Error(fmt.Sprintf(
			"Failed to write step result to log file: failed to marshal json: %v",
			err,
		))

		return
	}

	_, err = workflowMetaFile.Seek(0, 0)
	if err != nil {
		t.soloCtx.Logger.Error(fmt.Sprintf(
			"Failed to write step result to log file: failed to seek open meta file: %v",
			err,
		))

		return
	}

	if n, err := workflowMetaFile.Write(workflowMetaData); err != nil || n != len(workflowMetaData) {
		t.soloCtx.Logger.Error(fmt.Sprintf(
			"Failed to write workflow step meta output to file: %s: %v",
			metaPath,
			err,
		))
	}
}
