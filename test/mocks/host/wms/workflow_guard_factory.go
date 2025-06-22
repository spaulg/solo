package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/wms"
	"github.com/stretchr/testify/mock"
)

type MockWorkflowGuardFactory struct {
	mock.Mock
}

func (m *MockWorkflowGuardFactory) Build(workflowNames []workflowcommon.WorkflowName, containerNames []string) wms_types.WorkflowGuard {
	args := m.Called(workflowNames, containerNames)

	if g, ok := args.Get(0).(wms_types.WorkflowGuard); ok {
		return g
	} else {
		return nil
	}
}
