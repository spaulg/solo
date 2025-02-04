package container

import (
	"fmt"
	"github.com/spaulg/solo/internal/pkg/solo/context"
)

type OrchestratorFactory interface {
	Build(soloCtx *context.CliContext) (Orchestrator, error)
}

type DefaultOrchestratorFactory struct{}

func NewOrchestratorFactory() OrchestratorFactory {
	return &DefaultOrchestratorFactory{}
}

func (t *DefaultOrchestratorFactory) Build(soloCtx *context.CliContext) (Orchestrator, error) {
	switch soloCtx.Config.Orchestrator {
	case "docker":
		return NewDockerOrchestrator(soloCtx), nil

	default:
		return nil, fmt.Errorf("unsupported orchestrator %s", soloCtx.Config.Orchestrator)
	}
}
