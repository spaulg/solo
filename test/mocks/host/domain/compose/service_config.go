package compose

import (
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/mock"

	"github.com/spaulg/solo/internal/pkg/impl/host/domain"
)

type MockServiceConfig struct {
	mock.Mock
}

func (m *MockServiceConfig) GetServiceWorkflow(eventName string) domain.ServiceWorkflowConfig {
	args := m.Called(eventName)
	return args.Get(0).(domain.ServiceWorkflowConfig)
}

func (m *MockServiceConfig) GetConfig() types.ServiceConfig {
	args := m.Called()
	return args.Get(0).(types.ServiceConfig)
}

func (m *MockServiceConfig) ResolveContainerWorkingDirectory(cwd string) string {
	args := m.Called(cwd)
	return args.String(0)
}
