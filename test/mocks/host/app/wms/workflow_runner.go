package wms

import (
	"github.com/stretchr/testify/mock"

	"github.com/spaulg/solo/internal/pkg/impl/host/shared/wms"
)

type MockWorkflowRunner struct {
	mock.Mock
}

func (m *MockWorkflowRunner) RunWorkflow(workflowSession wms.WorkflowSession) (bool, error) {
	args := m.Called(workflowSession)
	return args.Bool(0), args.Error(1)
}
