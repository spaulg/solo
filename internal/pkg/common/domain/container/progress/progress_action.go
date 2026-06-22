package progress

type ComposeProgressAction int

const (
	Unknown ComposeProgressAction = iota
	Build
	Create
	Start
	Stop
	Remove
)

func (t ComposeProgressAction) String() string {
	switch t {
	case Build:
		return "Building"
	case Create:
		return "Creating"
	case Start:
		return "Starting"
	case Stop:
		return "Stopping"
	case Remove:
		return "Removing"
	default:
		return "Unknown"
	}
}
