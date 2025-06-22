package subscribers

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
	"github.com/spaulg/solo/internal/pkg/impl/host/wms"
	events_types "github.com/spaulg/solo/internal/pkg/types/host/events"
)

type LogWriterEventSubscriber struct {
	soloCtx *context.CliContext
	mu      sync.RWMutex
}

type StepLogMeta struct {
	ExitCode         uint8    `json:"exit_code"`
	Command          string   `json:"command"`
	Arguments        []string `json:"arguments"`
	WorkingDirectory string   `json:"working_directory"`
}

type WorkflowMeta map[string][]string

func NewLogWriterEventSubscriber(soloCtx *context.CliContext) events_types.Subscriber {
	return &LogWriterEventSubscriber{
		soloCtx: soloCtx,
	}
}

func (t *LogWriterEventSubscriber) Publish(event events_types.Event) {
	switch e := event.(type) {
	case *wms.WorkflowStepOutputEvent:
		t.writeStepOutput(e)

	case *wms.WorkflowStepCompleteEvent:
		t.writeStepResult(e)
	}
}

func (t *LogWriterEventSubscriber) writeStepOutput(e *wms.WorkflowStepOutputEvent) {
	if e.Stderr == "" && e.Stdout == "" {
		return
	}

	outputDirectory := path.Join(
		t.soloCtx.Project.GetStateDirectoryRoot(),
		"workflow-logs",
		t.soloCtx.TriggerDateTime.Format("2006-01-02T15-04-05.999999999Z"),
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
		stderrPath := path.Join(outputDirectory, e.StepId+".stderr")
		stderrFile, err := os.OpenFile(stderrPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			t.soloCtx.Logger.Error(fmt.Sprintf(
				"Failed to write step output to log file: failed to open stderr file: %s: %v",
				stderrPath,
				err,
			))

			return
		}

		defer func(stderrFile *os.File) {
			if err = stderrFile.Close(); err != nil {
				t.soloCtx.Logger.Error(fmt.Sprintf(
					"Failed to close step stderr output file: %v",
					err,
				))
			}
		}(stderrFile)

		if n, err := stderrFile.WriteString(e.Stderr); err != nil || n != len(e.Stderr) {
			t.soloCtx.Logger.Error(fmt.Sprintf(
				"Failed to write complete step stderr output to file: %s: %v",
				stderrPath,
				err,
			))
		}
	}

	if e.Stdout != "" {
		stdoutPath := path.Join(outputDirectory, e.StepId+".stdout")
		stdoutFile, err := os.OpenFile(stdoutPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			t.soloCtx.Logger.Error(fmt.Sprintf(
				"Failed to write step output to log file: failed to open stdout file: %s: %v",
				stdoutPath,
				err,
			))

			return
		}

		defer func(stdoutFile *os.File) {
			if err = stdoutFile.Close(); err != nil {
				t.soloCtx.Logger.Error(fmt.Sprintf(
					"Failed to close step stdout output file: %v",
					err,
				))
			}
		}(stdoutFile)

		if n, err := stdoutFile.WriteString(e.Stdout); err != nil || n != len(e.Stdout) {
			t.soloCtx.Logger.Error(fmt.Sprintf(
				"Failed to write complete step stdout output to file: %s: %v",
				stdoutPath,
				err,
			))
		}
	}
}

func (t *LogWriterEventSubscriber) writeStepResult(e *wms.WorkflowStepCompleteEvent) {
	outputDirectory := path.Join(
		t.soloCtx.Project.GetStateDirectoryRoot(),
		"workflow-logs",
		t.soloCtx.TriggerDateTime.Format("2006-01-02T15-04-05.999999999Z"),
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

	metaPath := path.Join(outputDirectory, e.StepId+".meta.json")
	metaJson := StepLogMeta{
		ExitCode:         e.ExitCode,
		Command:          e.Command,
		Arguments:        e.Arguments,
		WorkingDirectory: e.Cwd,
	}

	metaData, err := json.MarshalIndent(metaJson, "", "  ")
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

	workflowMeta[e.ServiceName] = append(workflowMeta[e.ServiceName], e.StepId)

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
