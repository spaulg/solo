package app

import (
	"github.com/spaulg/solo/internal/pkg/host/app/context"
	"github.com/spaulg/solo/internal/pkg/host/app/event_manager"
	"github.com/spaulg/solo/internal/pkg/host/infra/container"
)

func ProjectToolingFactory(soloCtx *context.CliContext) (*ProjectTooling, error) {
	// Event manager
	eventManager := event_manager.GetEventManagerInstance()

	// Container orchestrator factory
	orchestratorFactory := container.NewOrchestratorFactory(soloCtx.Logger, soloCtx.Config, soloCtx.Project, eventManager)

	projectTooling := NewProjectTooling(soloCtx, orchestratorFactory)

	return projectTooling, nil
}
