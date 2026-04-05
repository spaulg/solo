package wms

import (
	"github.com/stretchr/testify/mock"

	commonworkflow "github.com/spaulg/solo/internal/pkg/shared/domain/wms"
)

type MockWorkflowExecTracker struct {
	mock.Mock
}

func (m *MockWorkflowExecTracker) Save() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockWorkflowExecTracker) MarkExecuted(
	serviceName string,
	workflowName commonworkflow.WorkflowName,
) (bool, error) {
	args := m.Called(serviceName, workflowName)
	return args.Bool(0), args.Error(1)
}
