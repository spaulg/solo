package domain

type ExecutionEventRepository interface {
	EntityRepository[*ExecutionEvent]
}
