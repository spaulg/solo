package wf

import "iter"

type Definition interface {
	StepIterator() iter.Seq[Step]
}
