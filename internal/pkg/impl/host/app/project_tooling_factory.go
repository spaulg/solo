package app

import (
	"github.com/spaulg/solo/internal/pkg/impl/host/app/context"
	"github.com/spaulg/solo/internal/pkg/impl/host/app/events"
	"github.com/spaulg/solo/internal/pkg/impl/host/infra/container"
)

func ProjectToolingFactory(soloCtx *context.CliContext) (*ProjectTooling, error) {
	// Event manager
	eventManager := events.GetEventManagerInstance()

	// Container orchestrator factory
	orchestratorFactory := container.NewOrchestratorFactory(soloCtx, eventManager)

	projectTooling := NewProjectTooling(soloCtx, orchestratorFactory)

	return projectTooling, nil
}
