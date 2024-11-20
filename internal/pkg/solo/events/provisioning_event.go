package events

type EventType int

const (
	Finished EventType = iota
)

func (t EventType) String() string {
	return [...]string{"Finished"}[t]
}

type ProvisioningEvent struct {
	EventType EventType
	Service   string
	Status    int
}
