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
