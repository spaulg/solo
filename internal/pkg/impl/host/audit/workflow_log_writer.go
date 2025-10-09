package audit

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"sync"

	"github.com/spaulg/solo/internal/pkg/impl/host/context"
	events_types "github.com/spaulg/solo/internal/pkg/types/host/events"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/wms"
	logging2 "github.com/spaulg/solo/internal/pkg/types/host/wms/logging"
)

const workflowLogsPath = "workflow_logs"
const workflowEventMetaFile = "event.meta.json"

type WorkflowLogWriter struct {
	soloCtx         *context.CliContext
	mu              sync.RWMutex
	outputDirectory string
}

func NewWorkflowLogWriter(soloCtx *context.CliContext) logging2.WorkflowLogWriter {
	outputDirectory := path.Join(
		soloCtx.Project.GetStateDirectoryRoot(),
		workflowLogsPath,
		soloCtx.TriggerDateTime.Format("2006-01-02T15-04-05.999999999Z"),
	)

	return &WorkflowLogWriter{
		soloCtx:         soloCtx,
		outputDirectory: outputDirectory,
	}
}

func (t *WorkflowLogWriter) RecordEvent(callback func() error) error {
	eventFile := path.Join(t.outputDirectory, workflowEventMetaFile)
	workflowEvent := NewWorkflowEvent(eventFile, t.soloCtx.CommandPath, t.soloCtx.CommandArgs)
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

func (t *WorkflowLogWriter) Publish(event events_types.Event) {
	switch e := event.(type) {
	case *wms_types.WorkflowStepOutputEvent:
		t.writeStepOutput(e)

	case *wms_types.WorkflowStepCompleteEvent:
		t.writeStepResult(e)
	}
}

func (t *WorkflowLogWriter) writeStepOutput(e *wms_types.WorkflowStepOutputEvent) {
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

func (t *WorkflowLogWriter) appendStepOutputFile(
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

func (t *WorkflowLogWriter) writeStepResult(e *wms_types.WorkflowStepCompleteEvent) {
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

	// Workflow step meta file
	metaPath := path.Join(outputDirectory, e.StepID+".meta.json")
	metaJSON := NewStepLogMeta(e.ExitCode, e.Command, e.Arguments, e.Cwd)
	if err := metaJSON.Persist(metaPath); err != nil {
		t.soloCtx.Logger.Error(fmt.Sprintf(
			"Failed to write step result to log file: failed to persist meta file: %s: %v",
			metaPath,
			err,
		))
	}

	// Add workflow step to workflow meta file
	t.mu.Lock()
	defer t.mu.Unlock()

	workflowMetaPath := path.Join(outputDirectory, e.WorkflowName.String()+".meta.json")
	workflowMeta, err := LoadWorkflowMeta(workflowMetaPath)
	if err != nil {
		t.soloCtx.Logger.Error(fmt.Sprintf(
			"Failed to write step result to log file: failed to load meta file: %s: %v",
			metaPath,
			err,
		))

		return
	}

	defer (func() {
		if err = workflowMeta.Close(); err != nil {
			t.soloCtx.Logger.Error(fmt.Sprintf(
				"Failed to write step result to log file: failed to persist meta file: %s: %v",
				metaPath,
				err,
			))
		}
	})()

	workflowMeta.AppendStep(e.ContainerName, e.StepID)

	if err = workflowMeta.Persist(); err != nil {
		t.soloCtx.Logger.Error(fmt.Sprintf(
			"Failed to write step result to log file: failed to persist meta file: %s: %v",
			metaPath,
			err,
		))
	}
}
