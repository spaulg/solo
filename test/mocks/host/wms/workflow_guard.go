package wms

import (
	workflowcommon "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	events_types "github.com/spaulg/solo/internal/pkg/types/host/events"
	"github.com/stretchr/testify/mock"
)

type MockWorkflowGuard struct {
	mock.Mock
}

func (m *MockWorkflowGuard) Publish(event events_types.Event) {
	m.Called(event)
}

func (m *MockWorkflowGuard) Wait(callback func(container string, guardCallback func(name workflowcommon.WorkflowName) error) error) error {
	args := m.Called(callback)
	return args.Error(0)
}
