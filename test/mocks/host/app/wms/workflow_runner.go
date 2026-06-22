package wms

import (
	"github.com/stretchr/testify/mock"

	"github.com/spaulg/solo/internal/pkg/host/app/wms/wf"
)

type MockWorkflowRunner struct {
	mock.Mock
}

func (m *MockWorkflowRunner) RunWorkflow(workflowSession wf.Session) error {
	args := m.Called(workflowSession)
	return args.Error(0)
}
