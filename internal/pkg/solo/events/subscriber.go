package events

type Subscriber interface {
	Subscribe(eventManager Manager)
	Publish(event Event)
}
