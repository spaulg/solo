package container

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/spaulg/solo/internal/pkg/host/app/context"
	events_types "github.com/spaulg/solo/internal/pkg/shared/app/events"
	container_types "github.com/spaulg/solo/internal/pkg/shared/infra/container"
)

type OrchestratorFactory struct {
	soloCtx      *context.CliContext
	eventManager events_types.Manager
}

func NewOrchestratorFactory(
	soloCtx *context.CliContext,
	eventManager events_types.Manager,
) *OrchestratorFactory {
	return &OrchestratorFactory{
		soloCtx:      soloCtx,
		eventManager: eventManager,
	}
}

func (t *OrchestratorFactory) Build() (container_types.Orchestrator, error) {
	orchestrator, binaryPath, err := t.findOrchestrator()
	if err != nil {
		return nil, fmt.Errorf("failed to find orchestrator: %w", err)
	}

	switch orchestrator {
	case "docker":
		return NewDockerOrchestrator(t.soloCtx, t.eventManager, binaryPath), nil

	default:
		return nil, fmt.Errorf("unsupported orchestrator %s", orchestrator)
	}
}

func (t *OrchestratorFactory) findOrchestrator() (string, string, error) {
	for _, orchestrator := range t.soloCtx.Config.Orchestration.SearchOrder {
		orchestratorConfig, ok := t.soloCtx.Config.Orchestration.Orchestrators[orchestrator]
		if !ok {
			continue
		}

		if binaryPath, err := exec.LookPath(orchestratorConfig.Binary); err == nil {
			return orchestrator, binaryPath, nil
		}
	}

	return "", "", errors.New("an orchestrator binary could not be found in the configured search paths")
}
