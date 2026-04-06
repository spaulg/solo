package events

import (
	"github.com/stretchr/testify/mock"

	events3 "github.com/spaulg/solo/internal/pkg/shared/app/events"
)

type MockEventManager struct {
	mock.Mock
}

func (m *MockEventManager) Subscribe(eventSubscriber events3.Subscriber) {
	m.Called(eventSubscriber)
}

func (m *MockEventManager) Unsubscribe(eventSubscriber events3.Subscriber) {
	m.Called(eventSubscriber)
}

func (m *MockEventManager) Publish(data events3.Event) {
	m.Called(data)
}

func (m *MockEventManager) Wait() {
	m.Called()
}
