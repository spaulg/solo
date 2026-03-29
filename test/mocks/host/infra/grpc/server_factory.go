package grpc

import (
	"github.com/stretchr/testify/mock"

	wms_types "github.com/spaulg/solo/internal/pkg/types/host/app/wms"
	project_types "github.com/spaulg/solo/internal/pkg/types/host/domain/project"
	container_types "github.com/spaulg/solo/internal/pkg/types/host/infra/container"
	grpc_types "github.com/spaulg/solo/internal/pkg/types/host/infra/grpc"
)

type MockGRPCServerFactory struct {
	mock.Mock
}

func (m *MockGRPCServerFactory) Build(
	orchestrator container_types.Orchestrator,
	workflowExecutionTracker wms_types.WorkflowExecTracker,
	project project_types.Project,
	port int,
) (grpc_types.Server, error) {
	args := m.Called(orchestrator, workflowExecutionTracker, project, port)
	server := args.Get(0)

	if s, ok := server.(grpc_types.Server); ok {
		return s, args.Error(1)
	}

	return nil, args.Error(1)
}
