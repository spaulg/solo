package compose

import "github.com/compose-spec/compose-go/v2/types"

type WorkflowStep struct {
	Name             string  `yaml:"name"`
	Run              string  `yaml:"run"`
	Shell            *string `yaml:"shell"`
	WorkingDirectory *string `yaml:"working_dir"`
}

type ServiceWorkflowConfig struct {
	Timeout *types.Duration `yaml:"timeout"`
	Shell   *string         `yaml:"shell"`
	Steps   []WorkflowStep  `yaml:"steps"`
}

type ServiceWorkflows map[string]ServiceWorkflowConfig
