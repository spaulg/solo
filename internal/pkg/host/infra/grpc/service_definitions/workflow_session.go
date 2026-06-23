package service_definitions

import (
	"errors"
	"fmt"
	"log/slog"

	"google.golang.org/grpc"

	commonworkflow "github.com/spaulg/solo/internal/pkg/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/common/infra/grpc/services"
	"github.com/spaulg/solo/internal/pkg/host/domain"
	interceptors2 "github.com/spaulg/solo/internal/pkg/host/infra/grpc/interceptors"
	"github.com/spaulg/solo/internal/pkg/host/infra/grpc/service_definitions/wfsession"
)

type WorkflowSession struct {
	project             domain.Project
	workflowName        commonworkflow.WorkflowName
	server              grpc.BidiStreamingServer[services.WorkflowStreamRequest, services.WorkflowStreamResponse]
	workflowExecTracker WorkflowExecTracker
	orchestrator        ContainerImageWorkingDirectoryResolver

	serviceName       string
	containerName     string
	fullContainerName string
}

func NewWorkflowSession(
	logger *slog.Logger,
	project domain.Project,
	workflowName commonworkflow.WorkflowName,
	server grpc.BidiStreamingServer[services.WorkflowStreamRequest, services.WorkflowStreamResponse],
	workflowExecTracker WorkflowExecTracker,
	orchestrator ContainerImageWorkingDirectoryResolver,
) (*WorkflowSession, error) {
	ctx := server.Context()

	// Extract service name
	serviceNameContextValueName := interceptors2.ServiceName(interceptors2.ServiceNameContextValueName)
	serviceName, ok := ctx.Value(serviceNameContextValueName).(string)
	if !ok {
		logger.Error("Service name not found")
		return nil, fmt.Errorf("unauthorized")
	}

	// Extract container name
	containerNameContextValueName := interceptors2.ContainerName(interceptors2.ContainerNameContextValueName)
	containerName, ok := ctx.Value(containerNameContextValueName).(string)
	if !ok {
		logger.Error("Container name not found")
		return nil, fmt.Errorf("unauthorized")
	}

	// Extract full container name
	fullContainerNameContextValueName := interceptors2.ContainerName(interceptors2.FullContainerNameContextValueName)
	fullContainerName, ok := ctx.Value(fullContainerNameContextValueName).(string)
	if !ok {
		logger.Error("Full container name not found")
		return nil, fmt.Errorf("unauthorized")
	}

	return &WorkflowSession{
		project:             project,
		workflowName:        workflowName,
		server:              server,
		workflowExecTracker: workflowExecTracker,
		orchestrator:        orchestrator,
		serviceName:         serviceName,
		containerName:       containerName,
		fullContainerName:   fullContainerName,
	}, nil
}

func (t *WorkflowSession) GetWorkflowName() commonworkflow.WorkflowName {
	return t.workflowName
}

func (t *WorkflowSession) HasServiceWorkflowRun(
	serviceName string,
) (bool, error) {
	if !t.workflowName.IsServiceWorkflow() {
		return false, nil
	}

	isFirstExecution, err := t.workflowExecTracker.MarkExecuted(serviceName, t.workflowName)
	if err != nil {
		return false, fmt.Errorf("failed to mark service workflow as executed: %w", err)
	}

	return !isFirstExecution, nil
}

func (t *WorkflowSession) HasFirstContainerWorkflowRun() bool {
	if !t.workflowName.IsFirstContainerWorkflow() {
		return false
	}

	firstWorkflowCompleteContextValueName := interceptors2.FirstContainerComplete(t.workflowName)
	firstWorkflowComplete, firstWorkflowOk := t.server.Context().Value(firstWorkflowCompleteContextValueName).(string)

	return firstWorkflowOk && firstWorkflowComplete == "true"
}

func (t *WorkflowSession) GetServiceName() string {
	return t.serviceName
}

func (t *WorkflowSession) GetContainerName() string {
	return t.containerName
}

func (t *WorkflowSession) GetFullContainerName() string {
	return t.fullContainerName
}

func (t *WorkflowSession) GetWorkingDirectory() (string, error) {
	var err error

	serviceName := t.GetServiceName()
	service := t.project.Services().GetService(serviceName)
	serviceWorkingDirectory := service.GetConfig().WorkingDir

	if serviceWorkingDirectory == "" {
		// Project does not define a working directory for the service
		// Use orchestrator to lookup the working directory from the service image
		serviceWorkingDirectory, err = t.orchestrator.ResolveImageWorkingDirectory(serviceName)
		if err != nil {
			return "", err
		}
	}

	return serviceWorkingDirectory, err
}

func (t *WorkflowSession) RunCommand(request *wfsession.RunCommandRequest) error {
	return t.server.Send(&services.WorkflowStreamResponse{
		Action: services.WorkflowAction_RUN_COMMAND_ACTION,
		RunCommand: &services.WorkflowRunCommand{
			Command:          request.Command,
			Arguments:        request.Arguments,
			WorkingDirectory: request.WorkingDirectory,
		},
	})
}

func (t *WorkflowSession) RecvCommandResponse() (*wfsession.CommandResponse, error) {
	result, err := t.server.Recv()

	if err != nil {
		return nil, err
	}

	if result.Result != services.WorkflowResult_RUN_COMMAND_RESULT {
		return nil, errors.New("unknown result")
	}

	var exitCodePtr *uint8
	var exitCode uint8

	if result.RunCommandResult.ExitCode != nil {
		exitCode = uint8(*result.RunCommandResult.ExitCode) // nolint:gosec
		exitCodePtr = &exitCode
	}

	return &wfsession.CommandResponse{
		Stdout:   result.RunCommandResult.Stdout,
		Stderr:   result.RunCommandResult.Stderr,
		ExitCode: exitCodePtr,
	}, nil
}

func (t *WorkflowSession) MarkCompletion() error {
	return t.server.Send(&services.WorkflowStreamResponse{
		Action: services.WorkflowAction_COMPLETE_ACTION,
	})
}
