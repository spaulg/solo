package container

import (
	"github.com/stretchr/testify/mock"

	container_types "github.com/spaulg/solo/internal/pkg/shared/infra/container"
)

type MockOrchestratorFactory struct {
	mock.Mock
}

func (m *MockOrchestratorFactory) Build() (container_types.Orchestrator, error) {
	args := m.Called()
	orchestrator := args.Get(0)

	if o, ok := orchestrator.(container_types.Orchestrator); ok {
		return o, args.Error(1)
	}

	return nil, args.Error(1)
}
