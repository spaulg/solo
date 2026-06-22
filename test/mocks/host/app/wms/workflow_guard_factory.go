package wms

import (
	"github.com/stretchr/testify/mock"

	workflowcommon "github.com/spaulg/solo/internal/pkg/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/host/app/wms/wf"
)

type MockWorkflowGuardFactory struct {
	mock.Mock
}

func (m *MockWorkflowGuardFactory) Build(workflowNames []workflowcommon.WorkflowName, containerNames []string) wf.Guard {
	args := m.Called(workflowNames, containerNames)

	if g, ok := args.Get(0).(wf.Guard); ok {
		return g
	}

	return nil
}
