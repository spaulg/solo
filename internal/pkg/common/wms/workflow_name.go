package wms

import (
	"fmt"
)

type Name int

// nolint:gochecknoglobals
var WorkflowNames = []Name{
	Build,
	PreStart,
	PostStart,
	PreStop,
	PostStop,
	PreDestroy,
	PostDestroy,
}

func FromString(name string) (Name, error) {
	switch name {
	case "build":
		return Build, nil
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

func (c Name) String() string {
	switch c {
	case Build:
		return "build"
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
	Undefined Name = iota
	Build
	PreStart
	PostStart
	PreStop
	PostStop
	PreDestroy
	PostDestroy
)
