package container

import "github.com/spaulg/solo/internal/pkg/solo/context"

type OrchestratorFactory interface {
	Build(soloCtx *context.CliContext) Orchestrator
}

type DefaultOrchestratorFactory struct{}

func NewOrchestratorFactory() OrchestratorFactory {
	return &DefaultOrchestratorFactory{}
}

func (t *DefaultOrchestratorFactory) Build(soloCtx *context.CliContext) Orchestrator {
	return &DockerOrchestrator{
		projectDirectory: soloCtx.Project.GetDirectory(),
		composeFile:      soloCtx.Project.GetGeneratedComposeFilePath(),
	}
}
