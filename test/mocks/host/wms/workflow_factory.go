package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	project_types "github.com/spaulg/solo/internal/pkg/types/host/project"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/wms"
	"github.com/stretchr/testify/mock"
)

type MockWorkflowFactory struct {
	mock.Mock
}

func (m *MockWorkflowFactory) Make(
	project project_types.Project,
	service string,
	workflowName workflowcommon.WorkflowName,
) wms_types.Orchestrator {
	args := m.Called(project, service, workflowName)
	if o, ok := args.Get(0).(wms_types.Orchestrator); ok {
		return o
	} else {
		return nil
	}
}
