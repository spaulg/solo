package wms

import (
	"fmt"
)

type WorkflowName int

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

func WorkflowNameFromString(name string) (WorkflowName, error) {
	switch name {
	case "first_pre_start":
		return FirstPreStart, nil
	case "pre_start":
		return PreStart, nil
	case "post_start":
		return PostStart, nil
	case "pre_stop":
		return PreStop, nil
	case "post_stop":
		return PostStop, nil
	case "pre_destroy":
		return PreDestroy, nil
	case "post_destroy":
		return PostDestroy, nil
	default:
		return Undefined, fmt.Errorf("unknown workflow name: %s", name)
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
