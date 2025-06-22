package container

type OrchestratorFactory interface {
	Build() (Orchestrator, error)
}
