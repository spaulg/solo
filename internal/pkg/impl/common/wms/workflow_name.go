package wms

type WorkflowName int

const (
	Undefined WorkflowName = iota
	FirstPreStartContainer
	PreStartContainer
	PostStartContainer
	FirstPostStartContainer
	PreStopContainer
	PreDestroyContainer
)

// nolint:gochecknoglobals
var WorkflowNames = []WorkflowName{
	FirstPreStartContainer,
	PreStartContainer,
	PostStartContainer,
	FirstPostStartContainer,
	PreStopContainer,
	PreDestroyContainer,
}

func WorkflowNameFromString(name string) WorkflowName {
	switch name {
	case "first_pre_start_container":
		return FirstPreStartContainer
	case "pre_start_container":
		return PreStartContainer
	case "post_start_container":
		return PostStartContainer
	case "first_post_start_container":
		return FirstPostStartContainer
	case "pre_stop_container":
		return PreStopContainer
	case "pre_destroy_container":
		return PreDestroyContainer
	default:
		return Undefined
	}
}

func (c WorkflowName) String() string {
	switch c {
	case FirstPreStartContainer:
		return "first_pre_start_container"
	case PreStartContainer:
		return "pre_start_container"
	case PostStartContainer:
		return "post_start_container"
	case FirstPostStartContainer:
		return "first_post_start_container"
	case PreStopContainer:
		return "pre_stop_container"
	case PreDestroyContainer:
		return "pre_destroy_container"
	default:
		return "unknown"
	}
}
