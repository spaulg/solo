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
	Stdout string
	Stderr string
}

type WorkflowStepCompleteEvent struct {
	events.BaseEvent
	ExitCode uint8
}

type WorkflowCompleteEvent struct {
	events.BaseEvent
}

type WorkflowErrorEvent struct {
	events.BaseEvent
	Err error
}
