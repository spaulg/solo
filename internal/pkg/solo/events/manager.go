package events

import "sync"

type Manager interface {
	Subscribe(eventSubscriber Subscriber)
	Publish(data Event)
}

type DefaultManager struct {
	subscribers []Subscriber
}

// nolint:gochecknoglobals
var (
	eventManagerInstance Manager
	eventManagerOnce     sync.Once
)

func GetEventManagerInstance() Manager {
	eventManagerOnce.Do(func() {
		eventManagerInstance = NewDefaultEventManager()
	})

	return eventManagerInstance
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
