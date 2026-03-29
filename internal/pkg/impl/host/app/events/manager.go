package events

import (
	"sync"

	events_types "github.com/spaulg/solo/internal/pkg/types/host/app/events"
)

type Manager struct {
	subscribers map[events_types.Subscriber]chan events_types.Event
	mu          sync.RWMutex
	wg          sync.WaitGroup
}

// nolint:gochecknoglobals
var (
	eventManagerInstance events_types.Manager
	eventManagerOnce     sync.Once
)

func GetEventManagerInstance() events_types.Manager {
	eventManagerOnce.Do(func() {
		eventManagerInstance = NewEventManager()
	})

	return eventManagerInstance
}

func NewEventManager() events_types.Manager {
	return &Manager{
		subscribers: make(map[events_types.Subscriber]chan events_types.Event),
	}
}

func (t *Manager) Subscribe(eventSubscriber events_types.Subscriber) {
	t.mu.Lock()
	defer t.mu.Unlock()

	subscriberChannel := make(chan events_types.Event, 30)
	t.subscribers[eventSubscriber] = subscriberChannel

	go func(subscriber events_types.Subscriber) {
		for val := range subscriberChannel {
			subscriber.Publish(val)
			t.wg.Done()
		}
	}(eventSubscriber)
}

func (t *Manager) Unsubscribe(eventSubscriber events_types.Subscriber) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if ch, exists := t.subscribers[eventSubscriber]; exists {
		delete(t.subscribers, eventSubscriber)
		close(ch)
	}
}

func (t *Manager) Publish(event events_types.Event) {
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
