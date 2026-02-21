package audit

import (
	"fmt"
	"path"
	"sync"

	"github.com/spaulg/solo/internal/pkg/impl/host/app/context"
	"github.com/spaulg/solo/internal/pkg/impl/host/domain"
	events_types "github.com/spaulg/solo/internal/pkg/types/host/app/events"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/app/wms"
)

const auditLogsPath = "audit_logs"
const executionEventFile = "event.json"
const metaFileSuffix = ".meta.json"

type Auditor struct {
	soloCtx                       *context.CliContext
	mu                            sync.Mutex
	outputDirectory               string
	executionEventRepository      domain.ExecutionEventRepository
	workflowLogMetaRepository     domain.WorkflowLogMetaRepository
	workflowStepLogMetaRepository domain.WorkflowStepLogMetaRepository
	logWriter                     domain.LogWriter
}

func NewAuditor(
	soloCtx *context.CliContext,
	executionEventRepository domain.ExecutionEventRepository,
	workflowLogMetaRepository domain.WorkflowLogMetaRepository,
	workflowStepLogMetaRepository domain.WorkflowStepLogMetaRepository,
	logWriter domain.LogWriter,
) *Auditor {
	outputDirectory := path.Join(
		soloCtx.Project.GetStateDirectoryRoot(),
		auditLogsPath,
		soloCtx.TriggerDateTime.Format("2006-01-02T15-04-05.999999999Z"),
	)

	return &Auditor{
		soloCtx:                       soloCtx,
		outputDirectory:               outputDirectory,
		executionEventRepository:      executionEventRepository,
		workflowLogMetaRepository:     workflowLogMetaRepository,
		workflowStepLogMetaRepository: workflowStepLogMetaRepository,
		logWriter:                     logWriter,
	}
}

func (t *Auditor) RecordExecutionEvent(callback func() error) error {
	eventFile := path.Join(t.outputDirectory, executionEventFile)

	workflowEvent := domain.NewExecutionEvent(t.soloCtx.CommandPath, t.soloCtx.CommandArgs)
	if err := t.executionEventRepository.Save(eventFile, workflowEvent); err != nil {
		return fmt.Errorf("failed to record workflow event start: %w", err)
	}

	res := callback()
	workflowEvent.MarkComplete(res)

	if err := t.executionEventRepository.Save(eventFile, workflowEvent); err != nil {
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

	t.mu.Lock()
	defer t.mu.Unlock()

	filePath := path.Join(outputDirectory, e.StepID+metaFileSuffix)

	meta, err := t.workflowStepLogMetaRepository.Load(filePath)
	if err != nil {
		t.soloCtx.Logger.Error(fmt.Sprintf(
			"Failed to write step output to log file: %s: %v",
			filePath,
			err,
		))

		return
	}

	if meta == nil {
		meta = domain.NewWorkflowStepLogMeta()

		if err := t.workflowStepLogMetaRepository.Save(filePath, meta); err != nil {
			t.soloCtx.Logger.Error(fmt.Sprintf(
				"Failed to write workflow step meta file: %s: %v",
				filePath,
				err,
			))
		}
	}

	// Combined file
	combinedOutputPath := path.Join(outputDirectory, e.StepID+".out")
	if err := t.logWriter.Append(combinedOutputPath, []byte(e.Stderr)); err != nil {
		t.soloCtx.Logger.Error(fmt.Sprintf(
			"Failed to write step output to log file: %s: %v",
			combinedOutputPath,
			err,
		))
	}

	if err := t.logWriter.Append(combinedOutputPath, []byte(e.Stdout)); err != nil {
		t.soloCtx.Logger.Error(fmt.Sprintf(
			"Failed to write step output to log file: %s: %v",
			combinedOutputPath,
			err,
		))
	}

	// stderr file
	stderrPath := path.Join(outputDirectory, e.StepID+".stderr")
	if err := t.logWriter.Append(stderrPath, []byte(e.Stderr)); err != nil {
		t.soloCtx.Logger.Error(fmt.Sprintf(
			"Failed to write step output to log file: %s: %v",
			stderrPath,
			err,
		))
	}

	// stdout file
	stdoutPath := path.Join(outputDirectory, e.StepID+".stdout")
	if err := t.logWriter.Append(stdoutPath, []byte(e.Stdout)); err != nil {
		t.soloCtx.Logger.Error(fmt.Sprintf(
			"Failed to write step output to log file: %s: %v",
			stdoutPath,
			err,
		))
	}
}

func (t *Auditor) writeStepResult(e *wms_types.WorkflowStepCompleteEvent) {
	outputDirectory := path.Join(
		t.outputDirectory,
		e.WorkflowName.String(),
	)

	t.mu.Lock()
	defer t.mu.Unlock()

	// Workflow step meta file
	filePath := path.Join(outputDirectory, e.StepID+metaFileSuffix)
	metaJSON, err := t.workflowStepLogMetaRepository.Load(filePath)

	if err != nil {
		t.soloCtx.Logger.Error(fmt.Sprintf(
			"Failed to load workflow meta file: failed to load meta file: %s: %v",
			filePath,
			err,
		))

		return
	}

	if metaJSON == nil {
		metaJSON = domain.NewWorkflowStepLogMeta()
	}

	metaJSON.SetExecutionInfo(e.ExitCode, e.Command, e.Arguments, e.Cwd)

	if err := t.workflowStepLogMetaRepository.Save(filePath, metaJSON); err != nil {
		t.soloCtx.Logger.Error(fmt.Sprintf(
			"Failed to write workflow step meta file: failed to save meta file: %s: %v",
			filePath,
			err,
		))
	}

	workflowMetaPath := path.Join(outputDirectory, e.WorkflowName.String()+".meta.json")
	workflowMeta, err := t.workflowLogMetaRepository.Load(workflowMetaPath)
	if err != nil {
		t.soloCtx.Logger.Error(fmt.Sprintf(
			"Failed to load workflow meta file: failed to load meta file: %s: %v",
			workflowMetaPath,
			err,
		))

		return
	}

	if workflowMeta == nil {
		workflowMeta = domain.NewWorkflowLogMeta()
	}

	workflowMeta.AppendStep(e.ContainerName, e.StepID)

	if err = t.workflowLogMetaRepository.Save(workflowMetaPath, workflowMeta); err != nil {
		t.soloCtx.Logger.Error(fmt.Sprintf(
			"Failed to write workflow meta file: failed to save meta file: %s: %v",
			workflowMetaPath,
			err,
		))
	}
}
