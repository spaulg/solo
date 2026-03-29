package wms

import (
	"github.com/stretchr/testify/mock"

	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	wms_types "github.com/spaulg/solo/internal/pkg/types/host/app/wms"
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
