package compose

import "github.com/compose-spec/compose-go/v2/types"

type WorkflowStep struct {
	Name             string  `yaml:"name"`
	Run              string  `yaml:"run"`
	WorkingDirectory *string `yaml:"working_dir"`
}

type ServiceWorkflowConfig struct {
	Timeout *types.Duration `yaml:"timeout"`
	Steps   []WorkflowStep  `yaml:"steps"`
}

type ServiceWorkflows map[string]ServiceWorkflowConfig
