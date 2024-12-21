package events

type Subscriber interface {
	GetSubscribedEvents() []EventType
	Publish(eventType EventType, event *Event)
}
