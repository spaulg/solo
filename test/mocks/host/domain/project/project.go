package project

import (
	"time"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/stretchr/testify/mock"

	project_types "github.com/spaulg/solo/internal/pkg/types/host/domain/project"
	"github.com/spaulg/solo/internal/pkg/types/host/domain/project/compose"
)

type MockProject struct {
	mock.Mock
}

func (m *MockProject) ReloadWithAllProfilesEnabled() (project_types.Project, error) {
	args := m.Called()

	if p, ok := args.Get(0).(project_types.Project); ok {
		return p, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockProject) ResolveStateDirectory(relativePath string) string {
	args := m.Called(relativePath)
	return args.String(0)
}

func (m *MockProject) GetAllServicesStateDirectory() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProject) GetServiceStateDirectoryRoot() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProject) GetServiceStateDirectory(serviceName string) string {
	args := m.Called(serviceName)
	return args.String(0)
}

func (m *MockProject) GetServiceLogDirectory(serviceName string) string {
	args := m.Called(serviceName)
	return args.String(0)
}

func (m *MockProject) GetServiceMountDirectory(serviceName string) string {
	args := m.Called(serviceName)
	return args.String(0)
}

func (m *MockProject) GetStateDirectoryRoot() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProject) GetDirectory() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProject) GetFilePath() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProject) GetServiceWorkflow(serviceName string, eventName string) compose.ServiceWorkflowConfig {
	args := m.Called(serviceName, eventName)
	return args.Get(0).(compose.ServiceWorkflowConfig)
}

func (m *MockProject) GetGeneratedComposeFilePath() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProject) GetMaxWorkflowTimeout(eventName string) time.Duration {
	args := m.Called(eventName)
	return args.Get(0).(time.Duration)
}

func (m *MockProject) ContainerNames(serviceNames []string) ([]string, error) {
	args := m.Called(serviceNames)

	if s, ok := args.Get(0).([]string); ok {
		return s, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockProject) ProfilesOfServices(serviceNames []string) ([]string, error) {
	args := m.Called(serviceNames)

	if s, ok := args.Get(0).([]string); ok {
		return s, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockProject) ReloadWithProfiles(profiles []string) error {
	args := m.Called(profiles)
	return args.Error(0)
}

func (m *MockProject) Tools() compose.Tools {
	args := m.Called()
	if t, ok := args.Get(0).(compose.Tools); ok {
		return t
	}

	return nil
}

func (m *MockProject) Profiles() []string {
	args := m.Called()
	if s, ok := args.Get(0).([]string); ok {
		return s
	}

	return nil
}

func (m *MockProject) GetCompose() *types.Project {
	args := m.Called()

	if c, ok := args.Get(0).(*types.Project); ok {
		return c
	}

	return nil
}

func (m *MockProject) Services() compose.Services {
	args := m.Called()
	return args.Get(0).(compose.Services)
}

func (m *MockProject) HasService(serviceName string) bool {
	args := m.Called(serviceName)
	return args.Get(0).(bool)
}

func (m *MockProject) ServiceNames() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockProject) ExclusiveServiceNames() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockProject) MarshalYAML() ([]byte, error) {
	args := m.Called()

	if b, ok := args.Get(0).([]byte); ok {
		return b, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockProject) Name() string {
	args := m.Called()
	return args.String(0)
}
