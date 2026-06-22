package wf

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/host/app/event_manager/events"
)

type Guard interface {
	Publish(event events.Event)
	Wait(callback func(container string, guardCallback func(name workflowcommon.WorkflowName) error) error) error
}
