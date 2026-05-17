package events

import (
	"github.com/stretchr/testify/mock"

	events2 "github.com/spaulg/solo/internal/pkg/impl/host/app/event_manager/events"
)

type MockEventManager struct {
	mock.Mock
}

func (m *MockEventManager) Subscribe(eventSubscriber events2.Subscriber) {
	m.Called(eventSubscriber)
}

func (m *MockEventManager) Unsubscribe(eventSubscriber events2.Subscriber) {
	m.Called(eventSubscriber)
}

func (m *MockEventManager) Publish(data events2.Event) {
	m.Called(data)
}

func (m *MockEventManager) Wait() {
	m.Called()
}
