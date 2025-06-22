package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	events_types "github.com/spaulg/solo/internal/pkg/types/host/events"
)

type WorkflowGuard interface {
	Publish(event events_types.Event)
	Wait(callback func(container string, guardCallback func(name workflowcommon.WorkflowName) error) error) error
}
