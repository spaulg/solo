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

func (m *MockOrchestrator) ComposeUp(serviceNames []string) error {
	args := m.Called(serviceNames)
	return args.Error(0)
}

func (m *MockOrchestrator) ComposeStop(serviceNames []string) error {
	args := m.Called(serviceNames)
	return args.Error(0)
}

func (m *MockOrchestrator) ComposeDown(serviceNames []string) error {
	args := m.Called(serviceNames)
	return args.Error(0)
}

func (m *MockOrchestrator) ComposeForkAndExecute(serviceName string, index int, command string, arguments []string, workingDirectory string) error {
	args := m.Called(serviceName, index, command, arguments, workingDirectory)
	return args.Error(0)
}

func (m *MockOrchestrator) ForkAndExecute(containerName string, command string, arguments []string, workingDirectory string) error {
	args := m.Called(containerName, command, arguments, workingDirectory)
	return args.Error(0)
}

func (m *MockOrchestrator) StartCommand(containerName string, command []string) error {
	args := m.Called(containerName, command)
	return args.Error(0)
}

func (m *MockOrchestrator) RunCommand(containerName string, command []string) (string, error) {
	args := m.Called(containerName, command)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockOrchestrator) GetHostGatewayHostname() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *MockOrchestrator) ServicesStatus(serviceNames []string) (*container_types.ServiceStatus, error) {
	args := m.Called(serviceNames)
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

func (m *MockOrchestrator) ResolveContainerNameFromMetadata(md metadata.MD) (string, error) {
	args := m.Called(md)
	return args.String(0), args.Error(1)
}

func (m *MockOrchestrator) ResolveContainerNameFromServiceName(serviceName string, index int) (string, error) {
	args := m.Called(serviceName, index)
	return args.String(0), args.Error(1)
}

func (m *MockOrchestrator) ResolveImageWorkingDirectory(service string) (string, error) {
	args := m.Called(service)
	return args.String(0), args.Error(1)
}
