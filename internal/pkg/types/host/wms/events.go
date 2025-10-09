package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/wms"
)

type BaseWorkflowEvent struct {
	ServiceName       string
	ContainerName     string
	FullContainerName string
	WorkflowName      workflowcommon.WorkflowName
}

type WorkflowStartedEvent struct {
	BaseWorkflowEvent
}

type WorkflowStepStartedEvent struct {
	BaseWorkflowEvent
	StepId    string
	Name      string
	Command   string
	Arguments []string
	Cwd       string
	Shell     string
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
	Shell     string
	ExitCode  uint8
}

type WorkflowCompleteEvent struct {
	BaseWorkflowEvent
	Successful bool
}

type WorkflowSkippedEvent struct {
	BaseWorkflowEvent
	Successful bool
}

type WorkflowErrorEvent struct {
	BaseWorkflowEvent
	Err error
}
