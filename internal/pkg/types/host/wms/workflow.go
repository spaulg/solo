package wms

import "iter"

type Workflow interface {
	StepIterator() iter.Seq[Step]
}
