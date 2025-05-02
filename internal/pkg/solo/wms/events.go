package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/common/wms"
)

type BaseWorkflowEvent struct {
	ServiceName  string
	WorkflowName workflowcommon.WorkflowName
}

type WorkflowStartedEvent struct {
	BaseWorkflowEvent
}

type WorkflowStepStartedEvent struct {
	BaseWorkflowEvent
	Name string
}

type WorkflowStepOutputEvent struct {
	BaseWorkflowEvent
	StepId string
	Stdout string
	Stderr string
}

type WorkflowStepCompleteEvent struct {
	BaseWorkflowEvent
	StepId    string
	Command   string
	Arguments []string
	Cwd       string
	ExitCode  uint8
}

type WorkflowCompleteEvent struct {
	BaseWorkflowEvent
	Successful bool
}

type WorkflowErrorEvent struct {
	BaseWorkflowEvent
	Err error
}
