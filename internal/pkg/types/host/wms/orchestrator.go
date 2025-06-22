package wms

import "iter"

type Orchestrator interface {
	StepIterator() iter.Seq[Step]
}
