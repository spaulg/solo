package logging

import (
	"github.com/stretchr/testify/mock"

	"github.com/spaulg/solo/internal/pkg/types/host/events"
)

type MockWorkflowLogWriter struct {
	mock.Mock
}

func (m *MockWorkflowLogWriter) Publish(event events.Event) {
	m.Called(event)
}

func (m *MockWorkflowLogWriter) RecordEvent(callback func() error) error {
	_ = m.Called(callback)
	return callback()
}
