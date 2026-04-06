package wms

import (
	events_types "github.com/spaulg/solo/internal/pkg/shared/app/events"
	workflowcommon "github.com/spaulg/solo/internal/pkg/shared/domain/wms"
)

type WorkflowGuard interface {
	Publish(event events_types.Event)
	Wait(callback func(container string, guardCallback func(name workflowcommon.WorkflowName) error) error) error
}
