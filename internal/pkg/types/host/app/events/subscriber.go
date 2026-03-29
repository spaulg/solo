package events

type Subscriber interface {
	Publish(event Event)
}
