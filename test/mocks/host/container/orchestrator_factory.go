package container

import (
	container_types "github.com/spaulg/solo/internal/pkg/types/host/container"
	"github.com/stretchr/testify/mock"
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
