package app

import (
	"github.com/spaulg/solo/internal/pkg/host/infra/container"
)

type OrchestratorFactory interface {
	Build() (container.Orchestrator, error)
}
