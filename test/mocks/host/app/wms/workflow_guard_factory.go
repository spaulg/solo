package wms

import (
	"github.com/stretchr/testify/mock"

	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
	wms_types "github.com/spaulg/solo/internal/pkg/impl/host/app/wms/workflow"
)

type MockWorkflowGuardFactory struct {
	mock.Mock
}

func (m *MockWorkflowGuardFactory) Build(workflowNames []workflowcommon.WorkflowName, containerNames []string) wms_types.Guard {
	args := m.Called(workflowNames, containerNames)

	if g, ok := args.Get(0).(wms_types.Guard); ok {
		return g
	}

	return nil
}
