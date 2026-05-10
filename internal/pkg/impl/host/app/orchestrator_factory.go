package app

import (
	"github.com/spaulg/solo/internal/pkg/impl/host/infra/container"
)

type OrchestratorFactory interface {
	Build() (container.Orchestrator, error)
}
