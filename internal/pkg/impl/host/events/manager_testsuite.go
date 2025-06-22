package events

import (
	"github.com/stretchr/testify/suite"

	"github.com/spaulg/solo/test/mocks/host/events"
)

type MockEvent struct {
	Data string
}

type EventManagerTestSuite struct {
	suite.Suite

	subscriber1 *events.MockSubscriber
	subscriber2 *events.MockSubscriber
}

func (t *EventManagerTestSuite) SetupTest() {
	t.subscriber1 = &events.MockSubscriber{}
	t.subscriber2 = &events.MockSubscriber{}
}

func (t *EventManagerTestSuite) TestSingleton() {
	manager1 := GetEventManagerInstance()
	t.NotNil(manager1)

	manager2 := GetEventManagerInstance()
	t.NotNil(manager2)

	t.Equal(manager1, manager2)
}

func (t *EventManagerTestSuite) TestDefaultManager_SubscribeAndPublish() {
	manager := NewEventManager()
	manager.Subscribe(t.subscriber1)

	event := MockEvent{Data: "test event"}
	t.subscriber1.On("Publish", event)

	manager.Publish(event)
	manager.Wait()
}

func (t *EventManagerTestSuite) TestDefaultManager_Unsubscribe() {
	manager := NewEventManager()
	manager.Subscribe(t.subscriber1)
	manager.Unsubscribe(t.subscriber1)

	event := MockEvent{Data: "test event"}
	manager.Publish(event)
	manager.Wait()
}

func (t *EventManagerTestSuite) TestDefaultManager_MultipleSubscribers() {
	manager := NewEventManager()
	manager.Subscribe(t.subscriber1)
	manager.Subscribe(t.subscriber2)

	event := MockEvent{Data: "test event"}
	t.subscriber1.On("Publish", event)
	t.subscriber2.On("Publish", event)

	manager.Publish(event)
	manager.Wait()
}
