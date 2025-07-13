package project

import "github.com/compose-spec/compose-go/v2/types"

const ToolExtensionName = "x-tools"
const ServiceWorkflowExtensionName = "x-workflows"

type WorkflowStep struct {
	Name string  `yaml:"name"`
	Run  string  `yaml:"run"`
	Cwd  *string `yaml:"cwd"`
}

type ServiceWorkflowConfig struct {
	Timeout *types.Duration `yaml:"timeout"`
	Steps   []WorkflowStep  `yaml:"steps"`
}

type ToolConfig struct {
	Description      string `mapstructure:"description" yaml:"description"`
	Command          string `mapstructure:"command" yaml:"command"`
	Service          string `mapstructure:"service" yaml:"service"`
	WorkingDirectory string `mapstructure:"working_directory" yaml:"working_directory"`
}

type ServiceWorkflows map[string]ServiceWorkflowConfig

type Tools map[string]ToolConfig
