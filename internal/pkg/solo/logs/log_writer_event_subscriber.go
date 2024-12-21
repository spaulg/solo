package logs

import (
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spaulg/solo/internal/pkg/solo/events"
)

type LogWriterEventSubscriber struct {
	soloCtx *context.SoloContext
}

func NewLogWriterEventSubscriber(soloCtx *context.SoloContext) events.Subscriber {
	return &LogWriterEventSubscriber{
		soloCtx: soloCtx,
	}
}

func (t *LogWriterEventSubscriber) GetSubscribedEvents() []events.EventType {
	return []events.EventType{events.CommandProgress, events.CommandFinished}
}

func (t *LogWriterEventSubscriber) Publish(eventType events.EventType, event *events.Event) {
	t.soloCtx.Logger.Info("LogWriterEventSubscriber:Publish")

	if eventType == events.CommandProgress { // todo: CommandOutput
		// todo: implement write of command output and exit code to disk
	}
}
