package wms

import (
	"fmt"
	"sync"

	commonworkflow "github.com/spaulg/solo/internal/pkg/common/domain/wms"
	domain2 "github.com/spaulg/solo/internal/pkg/host/domain"
)

type WorkflowExecTracker struct {
	mu            sync.Mutex
	workflowTrace *domain2.WorkflowExecTrace
	repository    domain2.WorkflowExecTraceRepository
	filePath      string
}

func NewWorkflowExecTracker(
	filePath string,
	repository domain2.WorkflowExecTraceRepository,
) (*WorkflowExecTracker, error) {
	workflowTrace, err := repository.Load(filePath)
	if err != nil {
		return nil, err
	} else if workflowTrace == nil {
		workflowTrace = domain2.NewWorkflowExecTrace()
	}

	return &WorkflowExecTracker{
		filePath:      filePath,
		workflowTrace: workflowTrace,
		repository:    repository,
	}, nil
}

func (t *WorkflowExecTracker) Save() error {
	return t.repository.Save(t.filePath, t.workflowTrace)
}

func (t *WorkflowExecTracker) MarkExecuted(
	serviceName string,
	workflowName commonworkflow.WorkflowName,
) (bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	loaded := t.workflowTrace.MarkExecuted(serviceName, workflowName)

	// Save workflowMap after modification
	if err := t.Save(); err != nil {
		return loaded, fmt.Errorf("marked executed but failed to save workflow list: %w", err)
	}

	return loaded, nil
}

func (t *WorkflowExecTracker) Clear(serviceName []string, workflowNames []commonworkflow.WorkflowName) error {
	if len(serviceName) > 0 && len(workflowNames) > 0 {
		t.mu.Lock()
		defer t.mu.Unlock()

		t.workflowTrace.Clear(serviceName, workflowNames)

		if err := t.Save(); err != nil {
			return fmt.Errorf("failed to clear workflow(s) and save workflow list: %w", err)
		}
	}

	return nil
}
