package wf

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/common/domain/wms"
)

type BaseWorkflowEvent struct {
	ServiceName       string
	ContainerName     string
	FullContainerName string
	WorkflowName      workflowcommon.WorkflowName
}

type StartedEvent struct {
	BaseWorkflowEvent
}

type StepStartedEvent struct {
	BaseWorkflowEvent
	StepID    string
	Name      string
	Command   string
	Arguments []string
	Cwd       string
	Shell     string
}

type StepOutputEvent struct {
	BaseWorkflowEvent
	StepID string
	Stdout string
	Stderr string
}

type StepCompleteEvent struct {
	BaseWorkflowEvent
	StepID    string
	Command   string
	Arguments []string
	Cwd       string
	Shell     string
	ExitCode  uint8
}

type CompleteEvent struct {
	BaseWorkflowEvent
	Successful bool
}

type SkippedEvent struct {
	BaseWorkflowEvent
	Successful bool
}

type ErrorEvent struct {
	BaseWorkflowEvent
	Err error
}
