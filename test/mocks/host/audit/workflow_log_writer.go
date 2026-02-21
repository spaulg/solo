package audit

import (
	"github.com/stretchr/testify/mock"

	"github.com/spaulg/solo/internal/pkg/types/host/events"
)

type MockAuditor struct {
	mock.Mock
}

func (m *MockAuditor) Publish(event events.Event) {
	m.Called(event)
}

func (m *MockAuditor) RecordExecutionEvent(callback func() error) error {
	_ = m.Called(callback)
	return callback()
}
