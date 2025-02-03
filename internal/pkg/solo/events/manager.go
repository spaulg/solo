package events

import (
	"sync"
)

type Manager interface {
	Subscribe(eventSubscriber Subscriber) chan Event
	Unsubscribe(eventSubscriber Subscriber)
	Publish(data Event)
}

type DefaultManager struct {
	subscribers map[Subscriber]chan Event
	mu          sync.RWMutex
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
		subscribers: make(map[Subscriber]chan Event),
	}
}

func (t *DefaultManager) Subscribe(eventSubscriber Subscriber) chan Event {
	t.mu.Lock()
	defer t.mu.Unlock()

	subscriberChannel := make(chan Event, 30)
	t.subscribers[eventSubscriber] = subscriberChannel
	return subscriberChannel
}

func (t *DefaultManager) Unsubscribe(eventSubscriber Subscriber) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if ch, exists := t.subscribers[eventSubscriber]; exists {
		delete(t.subscribers, eventSubscriber)
		close(ch)
	}
}

func (t *DefaultManager) Publish(event Event) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, ch := range t.subscribers {
		go func(ch chan Event) {
			ch <- event
		}(ch)
	}
}
