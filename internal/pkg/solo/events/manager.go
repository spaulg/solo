package events

type Manager interface {
	Subscribe(eventSubscriber Subscriber)
	Publish(data Event)
}

type DefaultManager struct {
	subscribers []Subscriber
}

func NewDefaultEventManager() Manager {
	return &DefaultManager{
		subscribers: []Subscriber{},
	}
}

func (t *DefaultManager) Subscribe(eventSubscriber Subscriber) {
	t.subscribers = append(t.subscribers, eventSubscriber)
}

func (t *DefaultManager) Publish(event Event) {
	for _, subscriber := range t.subscribers {
		subscriber.Publish(event)
	}
}
