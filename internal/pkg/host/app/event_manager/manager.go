package event_manager

import (
	"sync"

	events2 "github.com/spaulg/solo/internal/pkg/host/app/event_manager/events"
)

type Manager struct {
	subscribers map[events2.Subscriber]chan events2.Event
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
		subscribers: make(map[events2.Subscriber]chan events2.Event),
	}
}

func (t *Manager) Subscribe(eventSubscriber events2.Subscriber) {
	t.mu.Lock()
	defer t.mu.Unlock()

	subscriberChannel := make(chan events2.Event, 30)
	t.subscribers[eventSubscriber] = subscriberChannel

	go func(subscriber events2.Subscriber) {
		for val := range subscriberChannel {
			subscriber.Publish(val)
			t.wg.Done()
		}
	}(eventSubscriber)
}

func (t *Manager) Unsubscribe(eventSubscriber events2.Subscriber) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if ch, exists := t.subscribers[eventSubscriber]; exists {
		delete(t.subscribers, eventSubscriber)
		close(ch)
	}
}

func (t *Manager) Publish(event events2.Event) {
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
