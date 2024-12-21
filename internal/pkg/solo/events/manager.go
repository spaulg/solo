package events

type Manager interface {
	Subscribe(eventSubscriber Subscriber)
	Publish(eventType EventType, data *Event)
}

type DefaultManager struct {
	subscribers map[EventType][]Subscriber
}

func NewDefaultEventManager() Manager {
	return &DefaultManager{
		subscribers: make(map[EventType][]Subscriber),
	}
}

func (t *DefaultManager) Subscribe(eventSubscriber Subscriber) {
	for _, eventType := range eventSubscriber.GetSubscribedEvents() {
		t.subscribers[eventType] = append(t.subscribers[eventType], eventSubscriber)
	}
}

func (t *DefaultManager) Publish(eventType EventType, event *Event) {
	for _, subscriber := range t.subscribers[eventType] {
		subscriber.Publish(eventType, event)
	}
}
