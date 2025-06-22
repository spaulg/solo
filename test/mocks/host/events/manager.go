package events

import (
	events_types "github.com/spaulg/solo/internal/pkg/types/host/events"
	"github.com/stretchr/testify/mock"
)

type MockEventManager struct {
	mock.Mock
}

func (m *MockEventManager) Subscribe(eventSubscriber events_types.Subscriber) {
	m.Called(eventSubscriber)
}

func (m *MockEventManager) Unsubscribe(eventSubscriber events_types.Subscriber) {
	m.Called(eventSubscriber)
}

func (m *MockEventManager) Publish(data events_types.Event) {
	m.Called(data)
}

func (m *MockEventManager) Wait() {
	m.Called()
}
