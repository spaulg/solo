package container

import (
	"fmt"
	"github.com/spaulg/solo/internal/pkg/solo/context"
	"github.com/spaulg/solo/internal/pkg/solo/events"
)

type OrchestratorFactory interface {
	Build() (Orchestrator, error)
}

type DefaultOrchestratorFactory struct {
	soloCtx      *context.CliContext
	eventManager events.Manager
}

func NewOrchestratorFactory(soloCtx *context.CliContext, eventManager events.Manager) OrchestratorFactory {
	return &DefaultOrchestratorFactory{
		soloCtx:      soloCtx,
		eventManager: eventManager,
	}
}

func (t *DefaultOrchestratorFactory) Build() (Orchestrator, error) {
	switch t.soloCtx.Config.Orchestrator {
	case "docker":
		return NewDockerOrchestrator(t.soloCtx, t.eventManager), nil

	default:
		return nil, fmt.Errorf("unsupported orchestrator %s", t.soloCtx.Config.Orchestrator)
	}
}
