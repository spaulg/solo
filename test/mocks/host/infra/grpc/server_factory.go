package grpc

import (
	"github.com/stretchr/testify/mock"

	"github.com/spaulg/solo/internal/pkg/impl/host/infra/grpc"
	wms_types "github.com/spaulg/solo/internal/pkg/impl/host/infra/grpc/service_definitions"
	project_types "github.com/spaulg/solo/internal/pkg/types/host/domain"
)

type MockGRPCServerFactory struct {
	mock.Mock
}

func (m *MockGRPCServerFactory) Build(
	orchestrator grpc.ContainerResolver,
	workflowExecutionTracker wms_types.WorkflowExecTracker,
	project project_types.Project,
	port int,
) (grpc.Server, error) {
	args := m.Called(orchestrator, workflowExecutionTracker, project, port)
	server := args.Get(0)

	if s, ok := server.(grpc.Server); ok {
		return s, args.Error(1)
	}

	return nil, args.Error(1)
}
