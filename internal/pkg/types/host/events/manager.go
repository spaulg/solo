package events

type Manager interface {
	Subscribe(eventSubscriber Subscriber)
	Unsubscribe(eventSubscriber Subscriber)
	Publish(data Event)
	Wait()
}
