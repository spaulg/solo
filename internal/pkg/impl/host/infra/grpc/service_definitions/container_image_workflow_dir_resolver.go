package service_definitions

type ContainerImageWorkingDirectoryResolver interface {
	ResolveImageWorkingDirectory(serviceName string) (string, error)
}
