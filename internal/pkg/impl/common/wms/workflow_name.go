package wms

type WorkflowName int

const (
	Undefined WorkflowName = iota
	FirstPreStartContainer
	FirstPreStartService
	PreStartContainer
	PreStartService
	PostStartContainer
	PostStartService
	FirstPostStartContainer
	FirstPostStartService
	PreStopContainer
	PreDestroyContainer
)

// nolint:gochecknoglobals
var WorkflowNames = []WorkflowName{
	FirstPreStartContainer,
	FirstPreStartService,
	PreStartContainer,
	PreStartService,
	PostStartContainer,
	PostStartService,
	FirstPostStartContainer,
	FirstPostStartService,
	PreStopContainer,
	PreDestroyContainer,
}

func WorkflowNameFromString(name string) WorkflowName {
	switch name {
	case "first_pre_start_container":
		return FirstPreStartContainer
	case "first_pre_start_service":
		return FirstPreStartService
	case "pre_start_container":
		return PreStartContainer
	case "pre_start_service":
		return PreStartService
	case "post_start_container":
		return PostStartContainer
	case "post_start_service":
		return PostStartService
	case "first_post_start_container":
		return FirstPostStartContainer
	case "first_post_start_service":
		return FirstPostStartService
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
	case FirstPreStartService:
		return "first_pre_start_service"
	case PreStartContainer:
		return "pre_start_container"
	case PreStartService:
		return "pre_start_service"
	case PostStartContainer:
		return "post_start_container"
	case PostStartService:
		return "post_start_service"
	case FirstPostStartContainer:
		return "first_post_start_container"
	case FirstPostStartService:
		return "first_post_start_service"
	case PreStopContainer:
		return "pre_stop_container"
	case PreDestroyContainer:
		return "pre_destroy_container"
	default:
		return "unknown"
	}
}

func (c WorkflowName) IsServiceWorkflow() bool {
	switch c {
	case FirstPreStartService, PreStartService, PostStartService, FirstPostStartService:
		return true
	default:
		return false
	}
}
