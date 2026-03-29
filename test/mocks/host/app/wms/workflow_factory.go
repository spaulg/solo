package wms

import (
	"github.com/stretchr/testify/mock"

	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	context_types "github.com/spaulg/solo/internal/pkg/impl/host/app/context"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/app/wms"
	container_types "github.com/spaulg/solo/internal/pkg/types/host/infra/container"
)

type MockWorkflowFactory struct {
	mock.Mock
}

func (m *MockWorkflowFactory) Make(
	soloCtx *context_types.CliContext,
	orchestrator container_types.Orchestrator,
	service string,
	workflowName workflowcommon.WorkflowName,
) (wms_types.Workflow, error) {
	args := m.Called(soloCtx, orchestrator, service, workflowName)
	if o, ok := args.Get(0).(wms_types.Workflow); ok {
		return o, args.Error(1)
	}

	return nil, args.Error(1)
}
