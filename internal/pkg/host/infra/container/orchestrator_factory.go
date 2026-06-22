package container

import (
	"errors"
	"fmt"
	"log/slog"
	"os/exec"

	"github.com/spaulg/solo/internal/pkg/host/app/event_manager/events"
	domain2 "github.com/spaulg/solo/internal/pkg/host/domain"
)

type OrchestratorFactory struct {
	logger       *slog.Logger
	config       *domain2.Config
	project      domain2.Project
	eventManager events.Manager
}

func NewOrchestratorFactory(
	logger *slog.Logger,
	config *domain2.Config,
	project domain2.Project,
	eventManager events.Manager,
) *OrchestratorFactory {
	return &OrchestratorFactory{
		logger:       logger,
		config:       config,
		project:      project,
		eventManager: eventManager,
	}
}

func (t *OrchestratorFactory) Build() (Orchestrator, error) {
	orchestrator, binaryPath, err := t.findOrchestrator()
	if err != nil {
		return nil, fmt.Errorf("failed to find orchestrator: %w", err)
	}

	switch orchestrator {
	case "docker":
		return NewDockerOrchestrator(t.logger, t.config, t.project, t.eventManager, binaryPath), nil

	default:
		return nil, fmt.Errorf("unsupported orchestrator %s", orchestrator)
	}
}

func (t *OrchestratorFactory) findOrchestrator() (string, string, error) {
	for _, orchestrator := range t.config.Orchestration.SearchOrder {
		orchestratorConfig, ok := t.config.Orchestration.Orchestrators[orchestrator]
		if !ok {
			continue
		}

		if binaryPath, err := exec.LookPath(orchestratorConfig.Binary); err == nil {
			return orchestrator, binaryPath, nil
		}
	}

	return "", "", errors.New("an orchestrator binary could not be found in the configured search paths")
}
