package subscribers

import (
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spaulg/solo/internal/pkg/solo/events"
	"github.com/spaulg/solo/internal/pkg/solo/wms"
)

type LogWriterEventSubscriber struct {
	receiver chan events.Event
	soloCtx  *context.CliContext
}

func NewLogWriterEventSubscriber(soloCtx *context.CliContext) events.Subscriber {
	return &LogWriterEventSubscriber{
		soloCtx: soloCtx,
	}
}

func (t *LogWriterEventSubscriber) Subscribe(eventManager events.Manager) {
	t.receiver = eventManager.Subscribe(t)

	go func() {
		for val := range t.receiver {
			t.Publish(val)
		}
	}()
}

func (t *LogWriterEventSubscriber) Publish(event events.Event) {
	switch event.(type) {
	case wms.WorkflowStepOutputEvent:
		// todo: implement write of command output and exit code to disk
	}
}
