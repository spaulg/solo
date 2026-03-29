package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
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
	StepID    string
	Name      string
	Command   string
	Arguments []string
	Cwd       string
	Shell     string
}

type WorkflowStepOutputEvent struct {
	BaseWorkflowEvent
	StepID string
	Stdout string
	Stderr string
}

type WorkflowStepCompleteEvent struct {
	BaseWorkflowEvent
	StepID    string
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
