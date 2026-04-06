package wms

import (
	"github.com/stretchr/testify/mock"

	wms_types "github.com/spaulg/solo/internal/pkg/shared/app/wms"
	workflowcommon "github.com/spaulg/solo/internal/pkg/shared/domain/wms"
)

type MockWorkflowGuardFactory struct {
	mock.Mock
}

func (m *MockWorkflowGuardFactory) Build(workflowNames []workflowcommon.WorkflowName, containerNames []string) wms_types.WorkflowGuard {
	args := m.Called(workflowNames, containerNames)

	if g, ok := args.Get(0).(wms_types.WorkflowGuard); ok {
		return g
	}

	return nil
}
