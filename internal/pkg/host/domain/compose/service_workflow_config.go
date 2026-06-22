package compose

import (
	"time"

	"github.com/compose-spec/compose-go/v2/types"

	"github.com/spaulg/solo/internal/pkg/host/domain"
)

type ServiceWorkflowConfig struct {
	TimeoutValue *types.Duration `mapstructure:"timeout" yaml:"timeout"`
	ShellValue   *string         `mapstructure:"shell" yaml:"shell"`
	StepsValue   []WorkflowStep  `mapstructure:"steps" yaml:"steps"`
}

func NewServiceWorkflowConfig(steps []domain.WorkflowStep, shell *string, timeout *types.Duration) ServiceWorkflowConfig {
	workflowSteps := make([]WorkflowStep, 0, len(steps))
	for _, step := range steps {
		if concreteStep, ok := step.(WorkflowStep); ok {
			workflowSteps = append(workflowSteps, concreteStep)
		}
	}

	return ServiceWorkflowConfig{
		StepsValue:   workflowSteps,
		ShellValue:   shell,
		TimeoutValue: timeout,
	}
}

func (t ServiceWorkflowConfig) Timeout() types.Duration {
	if t.TimeoutValue != nil {
		return *t.TimeoutValue
	}

	return types.Duration(60 * time.Second)
}

func (t ServiceWorkflowConfig) Shell() *string {
	return t.ShellValue
}

func (t ServiceWorkflowConfig) Steps() []domain.WorkflowStep {
	steps := make([]domain.WorkflowStep, 0, len(t.StepsValue))
	for _, step := range t.StepsValue {
		steps = append(steps, step)
	}

	return steps
}
