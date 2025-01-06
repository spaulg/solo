package messages

import (
	"github.com/spaulg/solo/internal/pkg/impl/host/container/progress"
	"github.com/spaulg/solo/internal/pkg/types/host/wms"
)

type ErrorMsg error

type ModelSizeMsg struct {
	Width  int
	Height int
}

type ComposeProgressMsg *progress.ComposeProgressEvent

type WorkflowStartedMsg *wms.WorkflowStartedEvent
type WorkflowStepStartedMsg *wms.WorkflowStepStartedEvent
type WorkflowStepOutputMsg *wms.WorkflowStepOutputEvent
type WorkflowStepCompleteMsg *wms.WorkflowStepCompleteEvent
type WorkflowCompleteMsg *wms.WorkflowCompleteEvent
