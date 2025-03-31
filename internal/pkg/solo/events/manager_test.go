package events

import (
	"github.com/stretchr/testify/suite"
	"sync"
	"testing"
)

type MockEvent struct {
	Data string
}

type MockSubscriber struct {
	receivedEvents []Event
	mu             sync.Mutex
}

type EventManagerTestSuite struct {
	suite.Suite
}

func TestEventManagerTestSuite(t *testing.T) {
	suite.Run(t, new(EventManagerTestSuite))
}

func (m *MockSubscriber) Publish(event Event) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.receivedEvents = append(m.receivedEvents, event)
}

func (m *MockSubscriber) GetReceivedEvents() []Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.receivedEvents
}

func (t *EventManagerTestSuite) TestSingleton() {
	manager1 := GetEventManagerInstance()
	t.NotNil(manager1)

	manager2 := GetEventManagerInstance()
	t.NotNil(manager2)

	t.Equal(manager1, manager2)
}

func (t *EventManagerTestSuite) TestDefaultManager_SubscribeAndPublish() {
	manager := NewDefaultEventManager()
	subscriber := &MockSubscriber{}
	manager.Subscribe(subscriber)

	event := MockEvent{Data: "test event"}
	manager.Publish(event)
	manager.Wait()

	receivedEvents := subscriber.GetReceivedEvents()
	t.Equal(1, len(receivedEvents))
	t.Equal(receivedEvents[0], event)
}

func (t *EventManagerTestSuite) TestDefaultManager_Unsubscribe() {
	manager := NewDefaultEventManager()
	subscriber := &MockSubscriber{}
	manager.Subscribe(subscriber)
	manager.Unsubscribe(subscriber)

	event := MockEvent{Data: "test event"}
	manager.Publish(event)
	manager.Wait()

	receivedEvents := subscriber.GetReceivedEvents()
	t.Zero(receivedEvents)
}

func (t *EventManagerTestSuite) TestDefaultManager_MultipleSubscribers() {
	manager := NewDefaultEventManager()
	subscriber1 := &MockSubscriber{}
	subscriber2 := &MockSubscriber{}
	manager.Subscribe(subscriber1)
	manager.Subscribe(subscriber2)

	event := MockEvent{Data: "test event"}
	manager.Publish(event)
	manager.Wait()

	receivedEvents1 := subscriber1.GetReceivedEvents()
	receivedEvents2 := subscriber2.GetReceivedEvents()

	t.Equal(1, len(receivedEvents1))
	t.Equal(1, len(receivedEvents2))

	t.Equal(receivedEvents1[0], event)
	t.Equal(receivedEvents2[0], event)
}
