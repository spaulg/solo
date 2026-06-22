package domain

import (
	commonworkflow "github.com/spaulg/solo/internal/pkg/common/domain/wms"
)

type WorkflowExecTrace struct {
	workflowMap  map[string]struct{}
	WorkflowList []string `json:"workflow_list"`
}

func NewWorkflowExecTrace() *WorkflowExecTrace {
	return &WorkflowExecTrace{
		workflowMap:  make(map[string]struct{}),
		WorkflowList: make([]string, 0),
	}
}

func (t *WorkflowExecTrace) ensureInitialized() {
	if t.workflowMap == nil {
		t.workflowMap = make(map[string]struct{}, len(t.WorkflowList))
		for _, key := range t.WorkflowList {
			t.workflowMap[key] = struct{}{}
		}
	}

	if t.WorkflowList == nil {
		t.WorkflowList = make([]string, 0)
	}
}

func (t *WorkflowExecTrace) MarkExecuted(serviceName string, workflowName commonworkflow.WorkflowName) bool {
	t.ensureInitialized()

	key := serviceName + ":" + workflowName.String()
	_, loaded := t.workflowMap[key]
	if !loaded {
		t.workflowMap[key] = struct{}{}
		t.WorkflowList = append(t.WorkflowList, key)
	}

	return !loaded
}

func (t *WorkflowExecTrace) Clear(serviceName []string, workflowNames []commonworkflow.WorkflowName) {
	t.ensureInitialized()

	for _, serviceName := range serviceName {
		for _, workflowName := range workflowNames {
			key := serviceName + ":" + workflowName.String()
			delete(t.workflowMap, key)
		}
	}

	t.WorkflowList = make([]string, 0)
	for key := range t.workflowMap {
		t.WorkflowList = append(t.WorkflowList, key)
	}
}

func (t *WorkflowExecTrace) Get() []string {
	return t.WorkflowList
}
