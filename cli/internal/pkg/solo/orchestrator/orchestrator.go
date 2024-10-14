package orchestrator

type Orchestrator interface {
	Start(projectDirectory string, composeFile string) error
	Stop(projectDirectory string, composeFile string) error
	Destroy(projectDirectory string, composeFile string) error
}

func BuildOrchestrator() Orchestrator {
	return &DockerComposeOrchestrator{}
}
