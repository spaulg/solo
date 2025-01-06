package subscribers

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spaulg/solo/internal/pkg/solo/bubbletea/messages"
	"github.com/spaulg/solo/internal/pkg/solo/container/progress"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spaulg/solo/internal/pkg/solo/events"
	"github.com/spaulg/solo/internal/pkg/solo/wms"
)

type EventBusToBubbleTeaBridge struct {
	soloCtx *context.CliContext
	program *tea.Program
}

func NewEventBusToBubbleTeaBridge(soloCtx *context.CliContext, program *tea.Program) events.Subscriber {
	return &EventBusToBubbleTeaBridge{
		soloCtx: soloCtx,
		program: program,
	}
}

func (m EventBusToBubbleTeaBridge) Publish(event events.Event) {
	switch e := event.(type) {
	case *progress.ComposeProgressEvent:
		m.soloCtx.Logger.Debug("Received ComposeProgressEvent")
		m.program.Send(messages.ComposeProgressMsg(e))

	case *wms.WorkflowStartedEvent:
		m.soloCtx.Logger.Debug("Received WorkflowStartedEvent")
		m.program.Send(messages.WorkflowStartedMsg(e))

	case *wms.WorkflowStepStartedEvent:
		m.soloCtx.Logger.Debug("Received WorkflowStepStartedEvent")
		m.program.Send(messages.WorkflowStepStartedMsg(e))

	case *wms.WorkflowStepOutputEvent:
		m.soloCtx.Logger.Debug("Received WorkflowStepOutputEvent")
		m.program.Send(messages.WorkflowStepOutputMsg(e))

	case *wms.WorkflowStepCompleteEvent:
		m.soloCtx.Logger.Debug("Received WorkflowStepCompleteEvent")
		m.program.Send(messages.WorkflowStepCompleteMsg(e))

	case *wms.WorkflowCompleteEvent:
		m.soloCtx.Logger.Debug("Received WorkflowCompleteEvent")
		m.program.Send(messages.WorkflowCompleteMsg(e))

	default:
		m.soloCtx.Logger.Error("Received unknown event")
	}
}
