package domain

type WorkflowExecTraceRepository interface {
	EntityRepository[*WorkflowExecTrace]
}
