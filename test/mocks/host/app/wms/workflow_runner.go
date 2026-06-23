package wms

import (
	"github.com/stretchr/testify/mock"

	"github.com/spaulg/solo/internal/pkg/host/infra/grpc/service_definitions/wfsession"
)

type MockWorkflowRunner struct {
	mock.Mock
}

func (m *MockWorkflowRunner) RunWorkflow(workflowSession wfsession.Session) error {
	args := m.Called(workflowSession)
	return args.Error(0)
}
