package progress

type ComposeProgressStatus int

const (
	UnknownProgress ComposeProgressStatus = iota
	InProgress
	Complete
)
