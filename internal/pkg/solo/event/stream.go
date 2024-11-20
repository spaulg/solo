package event

type Stream[E any] interface {
	Push(*E)
}
