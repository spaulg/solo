package workflow

import "iter"

type Definition interface {
	StepIterator() iter.Seq[Step]
}
