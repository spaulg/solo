package compose

import "iter"

type Services interface {
	ServiceConfigIterator() iter.Seq2[string, ServiceConfig]
	GetService(serviceName string) ServiceConfig
	HasService(serviceName string) bool
	ServiceNames() []string
	ExclusiveServiceNames() []string
	ContainerNames(serviceNames []string) ([]string, error)
	ProfilesOfServices(serviceNames []string) ([]string, error)
}
