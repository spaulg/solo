package event

type FifoStream[E any] struct{}

func NewStream[E any]() Stream[E] {
	return &FifoStream[E]{}
}

func (t *FifoStream[E]) Push(event *E) {
	// todo: place event in to a queue / channel to distribute to listeners
}
