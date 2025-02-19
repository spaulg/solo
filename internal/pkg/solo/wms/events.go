package wms

import "github.com/spaulg/solo/internal/pkg/solo/events"

type WorkflowStartedEvent struct {
	events.BaseEvent
}

type WorkflowStepStartedEvent struct {
	events.BaseEvent
	Name string
}

type WorkflowStepOutputEvent struct {
	events.BaseEvent
	StepId string
	Stdout string
	Stderr string
}

type WorkflowStepCompleteEvent struct {
	events.BaseEvent
	StepId    string
	Command   string
	Arguments []string
	Cwd       string
	ExitCode  uint8
}

type WorkflowCompleteEvent struct {
	events.BaseEvent
	Successful bool
}

type WorkflowErrorEvent struct {
	events.BaseEvent
	Err error
}
