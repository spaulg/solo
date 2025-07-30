package service_definitions

import (
	"fmt"
	"sync"

	commonworkflow "github.com/spaulg/solo/internal/pkg/impl/common/wms"
)

type WorkflowExecTracker struct {
	executedWorkflows sync.Map
}

func LoadWorkflowExecTracker() *WorkflowExecTracker {
	// todo: load any state out of storage

	return &WorkflowExecTracker{}
}

func (t *WorkflowExecTracker) Save() error {
	// todo: persist the current state of executed workflows
	return nil
}

func (t *WorkflowExecTracker) MarkExecuted(serviceName string, workflowName commonworkflow.WorkflowName) (bool, error) {
	_, loaded := t.executedWorkflows.LoadOrStore(serviceName + ":" + workflowName.String(), true)

	// Save state after modification
	if err := t.Save(); err != nil {
		return !loaded, fmt.Errorf("marked executed but failed to save state: %w", err)
	}

	return !loaded, nil
}
