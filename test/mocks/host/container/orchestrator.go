package container

import (
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"

	config_types "github.com/spaulg/solo/internal/pkg/types/host/config"
	container_types "github.com/spaulg/solo/internal/pkg/types/host/container"
	project_types "github.com/spaulg/solo/internal/pkg/types/host/project"
)

type MockOrchestrator struct {
	mock.Mock
}

func (m *MockOrchestrator) ComposeUp() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockOrchestrator) ComposeStop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockOrchestrator) ComposeDown() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockOrchestrator) Execute(containerName string, command []string) error {
	args := m.Called(containerName, command)
	return args.Error(0)
}

func (m *MockOrchestrator) GetHostGatewayHostname() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *MockOrchestrator) ServicesStatus() (*container_types.ServiceStatus, error) {
	args := m.Called()
	serviceStatus := args.Get(0)

	if s, ok := serviceStatus.(*container_types.ServiceStatus); ok {
		return s, args.Error(1)
	} else {
		return nil, args.Error(1)
	}
}

func (m *MockOrchestrator) ExportComposeConfiguration(config *config_types.Config, project project_types.Project) ([]byte, error) {
	args := m.Called(config, project)
	configBytes := args.Get(0)

	if b, ok := configBytes.([]byte); ok {
		return b, args.Error(1)
	} else {
		return nil, args.Error(1)
	}
}

func (m *MockOrchestrator) ResolveContainerNameFromMetadata(md metadata.MD) (*string, error) {
	args := m.Called(md)
	containerName := args.Get(0)

	if s, ok := containerName.(*string); ok {
		return s, args.Error(1)
	} else {
		return nil, args.Error(1)
	}
}
