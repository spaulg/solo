package subscribers

import (
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spaulg/solo/internal/pkg/solo/events"
	"github.com/spaulg/solo/internal/pkg/solo/wms"
)

type LogWriterEventSubscriber struct {
	soloCtx *context.SoloContext
}

func NewLogWriterEventSubscriber(soloCtx *context.SoloContext) events.Subscriber {
	return &LogWriterEventSubscriber{
		soloCtx: soloCtx,
	}
}

func (t *LogWriterEventSubscriber) Publish(event events.Event) {
	switch event.(type) {
	case wms.WorkflowStepOutputEvent:
		// todo: implement write of command output and exit code to disk
	}
}
