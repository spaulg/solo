package events

import (
	"sync"

	events3 "github.com/spaulg/solo/internal/pkg/shared/app/events"
)

type Manager struct {
	subscribers map[events3.Subscriber]chan events3.Event
	mu          sync.RWMutex
	wg          sync.WaitGroup
}

// nolint:gochecknoglobals
var (
	eventManagerInstance *Manager
	eventManagerOnce     sync.Once
)

func GetEventManagerInstance() *Manager {
	eventManagerOnce.Do(func() {
		eventManagerInstance = NewEventManager()
	})

	return eventManagerInstance
}

func NewEventManager() *Manager {
	return &Manager{
		subscribers: make(map[events3.Subscriber]chan events3.Event),
	}
}

func (t *Manager) Subscribe(eventSubscriber events3.Subscriber) {
	t.mu.Lock()
	defer t.mu.Unlock()

	subscriberChannel := make(chan events3.Event, 30)
	t.subscribers[eventSubscriber] = subscriberChannel

	go func(subscriber events3.Subscriber) {
		for val := range subscriberChannel {
			subscriber.Publish(val)
			t.wg.Done()
		}
	}(eventSubscriber)
}

func (t *Manager) Unsubscribe(eventSubscriber events3.Subscriber) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if ch, exists := t.subscribers[eventSubscriber]; exists {
		delete(t.subscribers, eventSubscriber)
		close(ch)
	}
}

func (t *Manager) Publish(event events3.Event) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, ch := range t.subscribers {
		t.wg.Add(1)
		ch <- event
	}
}

func (t *Manager) Wait() {
	t.wg.Wait()
}
