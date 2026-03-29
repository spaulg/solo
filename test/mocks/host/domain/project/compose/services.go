package compose

import (
	"iter"

	"github.com/stretchr/testify/mock"

	compose_types "github.com/spaulg/solo/internal/pkg/types/host/domain/project/compose"
)

type MockServices struct {
	mock.Mock
}

func (m *MockServices) ServiceConfigIterator() iter.Seq2[string, compose_types.ServiceConfig] {
	args := m.Called()
	return args.Get(0).(iter.Seq2[string, compose_types.ServiceConfig])
}

func (m *MockServices) GetService(serviceName string) compose_types.ServiceConfig {
	args := m.Called(serviceName)

	if s, ok := args.Get(0).(compose_types.ServiceConfig); ok {
		return s
	}

	return nil
}

func (m *MockServices) HasService(serviceName string) bool {
	args := m.Called(serviceName)
	return args.Bool(0)
}

func (m *MockServices) ServiceNames() []string {
	args := m.Called()
	if names, ok := args.Get(0).([]string); ok {
		return names
	}

	return nil
}

func (m *MockServices) ExclusiveServiceNames() []string {
	args := m.Called()
	if names, ok := args.Get(0).([]string); ok {
		return names
	}

	return nil
}

func (m *MockServices) ContainerNames(serviceNames []string) ([]string, error) {
	args := m.Called(serviceNames)
	if names, ok := args.Get(0).([]string); ok {
		return names, args.Error(1)
	}

	return nil, args.Error(1)
}

func (m *MockServices) ProfilesOfServices(serviceNames []string) ([]string, error) {
	args := m.Called(serviceNames)
	if profiles, ok := args.Get(0).([]string); ok {
		return profiles, args.Error(1)
	}

	return nil, args.Error(1)
}
