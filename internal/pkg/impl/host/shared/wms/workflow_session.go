package wms

import (
	commonworkflow "github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"
)

type RunCommandRequest struct {
	Command          string
	Arguments        []string
	WorkingDirectory string
}

type CommandResponse struct {
	Stdout   string
	Stderr   string
	ExitCode *uint8
}

type WorkflowSession interface {
	GetWorkflowName() commonworkflow.WorkflowName
	HasServiceWorkflowRun(serviceName string) (bool, error)
	HasFirstContainerWorkflowRun() bool
	GetServiceName() string
	GetContainerName() string
	GetFullContainerName() string
	GetWorkingDirectory() (string, error)

	RunCommand(*RunCommandRequest) error
	RecvCommandResponse() (*CommandResponse, error)
}
