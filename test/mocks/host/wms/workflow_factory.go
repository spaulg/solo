package wms

import (
	"github.com/stretchr/testify/mock"

	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	context_types "github.com/spaulg/solo/internal/pkg/impl/host/context"
	container_types "github.com/spaulg/solo/internal/pkg/types/host/container"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/wms"
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
	} else {
		return nil, args.Error(1)
	}
}
