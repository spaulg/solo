package wms

import (
	"github.com/stretchr/testify/mock"

	"github.com/spaulg/solo/internal/pkg/impl/host/app/wms/workflow"
)

type MockWorkflowRunner struct {
	mock.Mock
}

func (m *MockWorkflowRunner) RunWorkflow(workflowSession workflow.Session) error {
	args := m.Called(workflowSession)
	return args.Error(0)
}
