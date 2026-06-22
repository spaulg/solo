package grpc

import (
	"github.com/stretchr/testify/mock"

	"github.com/spaulg/solo/internal/pkg/host/domain"
	grpc2 "github.com/spaulg/solo/internal/pkg/host/infra/grpc"
	"github.com/spaulg/solo/internal/pkg/host/infra/grpc/service_definitions"
)

type MockGRPCServerFactory struct {
	mock.Mock
}

func (m *MockGRPCServerFactory) Build(
	orchestrator grpc2.ContainerResolver,
	workflowExecutionTracker service_definitions.WorkflowExecTracker,
	project domain.Project,
	port int,
) (grpc2.Server, error) {
	args := m.Called(orchestrator, workflowExecutionTracker, project, port)
	server := args.Get(0)

	if s, ok := server.(grpc2.Server); ok {
		return s, args.Error(1)
	}

	return nil, args.Error(1)
}
