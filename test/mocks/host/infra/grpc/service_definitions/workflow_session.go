package service_definitions

import (
	"github.com/stretchr/testify/mock"

	commonworkflow "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	shared_wms "github.com/spaulg/solo/internal/pkg/impl/host/shared/wms"
)

type MockWorkflowSession struct {
	mock.Mock
}

func (m *MockWorkflowSession) GetWorkflowName() commonworkflow.WorkflowName {
	args := m.Called()
	return args.Get(0).(commonworkflow.WorkflowName)
}

func (m *MockWorkflowSession) HasServiceWorkflowRun(serviceName string) (bool, error) {
	args := m.Called(serviceName)
	return args.Bool(0), args.Error(1)
}

func (m *MockWorkflowSession) HasFirstContainerWorkflowRun() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockWorkflowSession) GetServiceName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockWorkflowSession) GetContainerName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockWorkflowSession) GetFullContainerName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockWorkflowSession) GetWorkingDirectory() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockWorkflowSession) RunCommand(request *shared_wms.RunCommandRequest) error {
	args := m.Called(request)
	return args.Error(0)
}

func (m *MockWorkflowSession) RecvCommandResponse() (*shared_wms.CommandResponse, error) {
	args := m.Called()
	if r, ok := args.Get(0).(*shared_wms.CommandResponse); ok {
		return r, args.Error(1)
	}

	return nil, args.Error(1)
}
