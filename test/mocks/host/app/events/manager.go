package events

import (
	"github.com/stretchr/testify/mock"

	"github.com/spaulg/solo/internal/pkg/types/host/app/events"
)

type MockEventManager struct {
	mock.Mock
}

func (m *MockEventManager) Subscribe(eventSubscriber events.Subscriber) {
	m.Called(eventSubscriber)
}

func (m *MockEventManager) Unsubscribe(eventSubscriber events.Subscriber) {
	m.Called(eventSubscriber)
}

func (m *MockEventManager) Publish(data events.Event) {
	m.Called(data)
}

func (m *MockEventManager) Wait() {
	m.Called()
}
