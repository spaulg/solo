package grpc

import (
	"github.com/stretchr/testify/mock"

	"github.com/spaulg/solo/internal/pkg/impl/host/domain"
	"github.com/spaulg/solo/internal/pkg/impl/host/infra/grpc"
	"github.com/spaulg/solo/internal/pkg/impl/host/infra/grpc/service_definitions"
)

type MockGRPCServerFactory struct {
	mock.Mock
}

func (m *MockGRPCServerFactory) Build(
	orchestrator grpc.ContainerResolver,
	workflowExecutionTracker service_definitions.WorkflowExecTracker,
	project domain.Project,
	port int,
) (grpc.Server, error) {
	args := m.Called(orchestrator, workflowExecutionTracker, project, port)
	server := args.Get(0)

	if s, ok := server.(grpc.Server); ok {
		return s, args.Error(1)
	}

	return nil, args.Error(1)
}
