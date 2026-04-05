package app

import (
	"github.com/spaulg/solo/internal/pkg/shared/infra/container"
)

type OrchestratorFactory interface {
	Build() (container.Orchestrator, error)
}
