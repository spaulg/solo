package compose

import (
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/mock"

	compose_types "github.com/spaulg/solo/internal/pkg/types/host/project/compose"
)

type MockServiceConfig struct {
	mock.Mock
}

func (m *MockServiceConfig) GetServiceWorkflow(eventName string) compose_types.ServiceWorkflowConfig {
	args := m.Called(eventName)
	return args.Get(0).(compose_types.ServiceWorkflowConfig)
}

func (m *MockServiceConfig) GetConfig() types.ServiceConfig {
	args := m.Called()
	return args.Get(0).(types.ServiceConfig)
}

func (t *MockServiceConfig) ResolveContainerWorkingDirectory(cwd string) string {
	args := t.Called(cwd)
	return args.String(0)
}
