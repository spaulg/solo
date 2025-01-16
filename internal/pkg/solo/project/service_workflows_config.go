package project

import "github.com/compose-spec/compose-go/v2/types"

const ServiceWorkflowExtensionName = "x-workflows"

type WorkflowStep struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
	Cwd     string `yaml:"cwd"`
}

type ServiceWorkflowConfig struct {
	Timeout *types.Duration `yaml:"timeout"`
	Steps   []WorkflowStep  `yaml:"steps"`
}

type ServiceWorkflows map[string]ServiceWorkflowConfig

func NewServiceWorkflows() ServiceWorkflows {
	return make(ServiceWorkflows)
}
