package messages

import (
	"github.com/spaulg/solo/internal/pkg/solo/container/progress"
	"github.com/spaulg/solo/internal/pkg/solo/wms"
)

type ErrorMsg error

type ComposeProgressMsg *progress.ComposeProgressEvent

type WorkflowStartedMsg *wms.WorkflowStartedEvent
type WorkflowStepStartedMsg *wms.WorkflowStepStartedEvent
type WorkflowStepOutputMsg *wms.WorkflowStepOutputEvent
type WorkflowStepCompleteMsg *wms.WorkflowStepCompleteEvent
type WorkflowCompleteMsg *wms.WorkflowCompleteEvent
