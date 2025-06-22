package container

import (
	"fmt"

	"github.com/spaulg/solo/internal/pkg/impl/host/context"
	container_types "github.com/spaulg/solo/internal/pkg/types/host/container"
	events_types "github.com/spaulg/solo/internal/pkg/types/host/events"
)

type OrchestratorFactory struct {
	soloCtx      *context.CliContext
	eventManager events_types.Manager
}

func NewOrchestratorFactory(soloCtx *context.CliContext, eventManager events_types.Manager) container_types.OrchestratorFactory {
	return &OrchestratorFactory{
		soloCtx:      soloCtx,
		eventManager: eventManager,
	}
}

func (t *OrchestratorFactory) Build() (container_types.Orchestrator, error) {
	orchestrator := t.soloCtx.Config.Orchestrator

	switch orchestrator {
	case "docker":
		return NewDockerOrchestrator(t.soloCtx, t.eventManager), nil

	default:
		return nil, fmt.Errorf("unsupported orchestrator %s", orchestrator)
	}
}
