package wms

import (
	"github.com/stretchr/testify/mock"

	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	container_types "github.com/spaulg/solo/internal/pkg/types/host/container"
	project_types "github.com/spaulg/solo/internal/pkg/types/host/project"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/wms"
)

type MockWorkflowFactory struct {
	mock.Mock
}

func (m *MockWorkflowFactory) Make(
	project project_types.Project,
	orchestrator container_types.Orchestrator,
	service string,
	workflowName workflowcommon.WorkflowName,
) (wms_types.Orchestrator, error) {
	args := m.Called(project, orchestrator, service, workflowName)
	if o, ok := args.Get(0).(wms_types.Orchestrator); ok {
		return o, args.Error(1)
	} else {
		return nil, args.Error(1)
	}
}
