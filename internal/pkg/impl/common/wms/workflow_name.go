package wms

type WorkflowName int

const (
	Undefined WorkflowName = iota
	FirstPreStart
	PreStart
	PostStart
	PreStop
	PostStop
	PreDestroy
	PostDestroy
)

// nolint:gochecknoglobals
var WorkflowNames = []WorkflowName{
	FirstPreStart,
	PreStart,
	PostStart,
	PreStop,
	PostStop,
	PreDestroy,
	PostDestroy,
}

func WorkflowNameFromString(name string) WorkflowName {
	switch name {
	case "first_pre_start":
		return FirstPreStart
	case "pre_start":
		return PreStart
	case "post_start":
		return PostStart
	case "pre_stop":
		return PreStop
	case "post_stop":
		return PostStop
	case "pre_destroy":
		return PreDestroy
	case "post_destroy":
		return PostDestroy
	default:
		return Undefined
	}
}

func (c WorkflowName) String() string {
	switch c {
	case FirstPreStart:
		return "first_pre_start"
	case PreStart:
		return "pre_start"
	case PostStart:
		return "post_start"
	case PreStop:
		return "pre_stop"
	case PostStop:
		return "post_stop"
	case PreDestroy:
		return "pre_destroy"
	case PostDestroy:
		return "post_destroy"
	default:
		return "unknown"
	}
}
