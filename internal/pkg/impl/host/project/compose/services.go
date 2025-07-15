package compose

import (
	"fmt"
	"iter"

	"github.com/compose-spec/compose-go/v2/types"

	compose_types "github.com/spaulg/solo/internal/pkg/types/host/project/compose"
)

type Services struct {
	compose *types.Project
}

func NewServices(compose *types.Project) compose_types.Services {
	return &Services{
		compose: compose,
	}
}

func (t *Services) GetService(serviceName string) compose_types.ServiceConfig {
	return NewServiceConfig(t.compose.Services[serviceName])
}

func (t *Services) ServiceConfigIterator() iter.Seq2[string, compose_types.ServiceConfig] {
	return func(yield func(string, compose_types.ServiceConfig) bool) {
		for name, config := range t.compose.Services {
			yield(name, NewServiceConfig(config))
		}
	}
}

func (t *Services) HasService(serviceName string) bool {
	_, exists := t.compose.Services[serviceName]
	return exists
}

func (t *Services) ServiceNames() []string {
	return t.compose.ServiceNames()
}

func (t *Services) ExclusiveServiceNames() []string {
	if len(t.compose.Profiles) == 1 && t.compose.Profiles[0] == "*" {
		return t.ServiceNames()
	}

	var exclusiveNames []string

	for _, service := range t.compose.Services {
		if len(service.Profiles) == 0 {
			continue
		}

		exclusiveNames = append(exclusiveNames, service.Name)
	}

	return exclusiveNames
}

// todo: this should be moved to the orchestrator as the format is orchestrator dependent
func (t *Services) ContainerNames(serviceNames []string) ([]string, error) {
	var containerNames []string

	if err := t.compose.ForEachService(serviceNames, func(name string, service *types.ServiceConfig) error {
		replicas := 1

		if service.Deploy != nil && service.Deploy.Replicas != nil {
			replicas = *service.Deploy.Replicas
		}

		if len(service.ContainerName) > 0 && replicas == 1 {
			// single container with a name defined by the container_name option
			containerNames = append(containerNames, service.ContainerName)
		} else {
			// one or more containers defined by the format {project}-{service}-{number}
			// consider moving this format to the orchestrator
			for i := 1; i <= replicas; i++ {
				containerName := fmt.Sprintf("%s-%s-%d", t.compose.Name, name, i)
				containerNames = append(containerNames, containerName)
			}
		}

		return nil
	}, types.IncludeDependencies); err != nil {
		return nil, err
	}

	return containerNames, nil
}

func (t *Services) ProfilesOfServices(serviceNames []string) ([]string, error) {
	var profileNames []string
	var profileNameMap = make(map[string]bool)

	for _, serviceName := range serviceNames {
		service, err := t.compose.GetService(serviceName)
		if err != nil {
			return nil, err
		}

		for _, profile := range service.Profiles {
			if profile == "*" {
				continue
			}

			if _, exists := profileNameMap[profile]; !exists {
				profileNames = append(profileNames, profile)
				profileNameMap[profile] = true
			}
		}
	}

	return profileNames, nil
}
