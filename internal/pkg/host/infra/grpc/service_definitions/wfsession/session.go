package wfsession

import (
	commonworkflow "github.com/spaulg/solo/internal/pkg/common/domain/wms"
)

type Session interface {
	GetWorkflowName() commonworkflow.WorkflowName
	HasServiceWorkflowRun(serviceName string) (bool, error)
	HasFirstContainerWorkflowRun() bool
	GetServiceName() string
	GetContainerName() string
	GetFullContainerName() string
	GetWorkingDirectory() (string, error)

	RunCommand(*RunCommandRequest) error
	RecvCommandResponse() (*CommandResponse, error)
	MarkCompletion() error
}
