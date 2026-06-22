package wms

import (
	"github.com/stretchr/testify/mock"

	workflowcommon "github.com/spaulg/solo/internal/pkg/common/domain/wms"
	events_types "github.com/spaulg/solo/internal/pkg/host/app/event_manager/events"
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
