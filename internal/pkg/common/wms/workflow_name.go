package wms

type Name int

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
	default:
		return "unknown"
	}
}

const (
	Build Name = iota
	PreStart
	PostStart
	PreStop
)
