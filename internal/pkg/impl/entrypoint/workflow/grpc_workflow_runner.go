package workflow

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/spaulg/solo/internal/pkg/impl/common/grpc/services"
	commonworkflow "github.com/spaulg/solo/internal/pkg/impl/common/wms"
	entrypointcontext "github.com/spaulg/solo/internal/pkg/impl/entrypoint/context"
	grpc_credentials_types "github.com/spaulg/solo/internal/pkg/types/entrypoint/grpc/credentials"
	workflow_types "github.com/spaulg/solo/internal/pkg/types/entrypoint/workflow"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const FirstPreStartCompleteMetadataKey = "first_pre_start_complete"

type GrpcWorkflowRunner struct {
	entrypointCtx  *entrypointcontext.EntrypointContext
	conn           *grpc.ClientConn
	workflowClient services.WorkflowClient
	metadataState  *MetadataState
}

type WorkflowStream grpc.BidiStreamingClient[services.WorkflowStreamRequest, services.WorkflowStreamResponse]

func NewGrpcWorkflowRunner(
	entrypointCtx *entrypointcontext.EntrypointContext,
	credentialsBuilder grpc_credentials_types.Builder,
	workflowServerHost string,
	metadataState *MetadataState,
) (workflow_types.WorkflowRunner, error) {
	entrypointCtx.Logger.Info("Connect to grpc server")

	creds, err := credentialsBuilder.Build()
	if err != nil {
		return nil, err
	}

	conn, err := grpc.NewClient(workflowServerHost, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, err
	}

	entrypointCtx.Logger.Info("Creating new service client")
	client := services.NewWorkflowClient(conn)

	return &GrpcWorkflowRunner{
		entrypointCtx:  entrypointCtx,
		conn:           conn,
		workflowClient: client,
		metadataState:  metadataState,
	}, nil
}

func (t *GrpcWorkflowRunner) Execute(workflowName commonworkflow.WorkflowName) error {
	stream, err := t.buildStream(workflowName)
	if err != nil {
		return err
	}

	defer func(stream WorkflowStream) {
		err := stream.CloseSend()
		if err != nil {
			t.entrypointCtx.Logger.Error("Failed to close GRPC stream")
		}
	}(stream)

	for {
		instruction, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		switch instruction.Action {
		case services.WorkflowAction_RUN_COMMAND_ACTION:
			t.entrypointCtx.Logger.Info(fmt.Sprintf("Running command: %s %v\n", instruction.RunCommand.Command, instruction.RunCommand.Arguments))

			exitCode, err := t.execute(
				instruction.RunCommand.Command,
				instruction.RunCommand.Arguments,
				instruction.RunCommand.WorkingDirectory,
				func(stdout string, stderr string) error {
					t.entrypointCtx.Logger.Info(fmt.Sprintf("%s\n", stdout))
					t.entrypointCtx.Logger.Info(fmt.Sprintf("%s\n", stderr))

					if err := stream.Send(&services.WorkflowStreamRequest{
						Result: services.WorkflowResult_RUN_COMMAND_RESULT,
						RunCommandResult: &services.WorkflowRunResult{
							Stdout: stdout,
							Stderr: stderr,
						},
					}); err != nil {
						return err
					}

					return nil
				},
			)

			if err != nil {
				return err
			}

			if err := stream.Send(&services.WorkflowStreamRequest{
				Result: services.WorkflowResult_RUN_COMMAND_RESULT,
				RunCommandResult: &services.WorkflowRunResult{
					ExitCode: &exitCode,
				},
			}); err != nil {
				return err
			}

		case services.WorkflowAction_COMPLETE_ACTION:
			if workflowName == commonworkflow.FirstPreStart {
				t.metadataState.Set(FirstPreStartCompleteMetadataKey, "true")
			}

		default:
			return errors.New("unknown action received from workflow stream")
		}
	}

	if err := t.metadataState.SaveToFile(); err != nil {
		return err
	}

	return nil
}

func (t *GrpcWorkflowRunner) Close() error {
	return t.conn.Close()
}

func (t *GrpcWorkflowRunner) buildStream(workflowName commonworkflow.WorkflowName) (WorkflowStream, error) {
	ctx := metadata.NewOutgoingContext(context.Background(), t.metadataState.ExportToGrpcMetadata())

	switch workflowName {
	case commonworkflow.FirstPreStart:
		return t.workflowClient.FirstPreStartWorkflowStream(ctx)
	case commonworkflow.PreStart:
		return t.workflowClient.PreStartWorkflowStream(ctx)
	case commonworkflow.PostStart:
		return t.workflowClient.PostStartWorkflowStream(ctx)
	case commonworkflow.PreStop:
		return t.workflowClient.PreStopWorkflowStream(ctx)
	case commonworkflow.PreDestroy:
		return t.workflowClient.PreDestroyWorkflowStream(ctx)
	default:
		return nil, errors.New("invalid wms name")
	}
}

func (t *GrpcWorkflowRunner) execute(
	command string,
	arguments []string,
	workingDirectory string,
	streamOutput func(stdout string, stderr string) error,
) (uint32, error) {
	cmd := exec.Command(command, arguments...)
	cmd.Dir = workingDirectory

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.entrypointCtx.Logger.Error(fmt.Sprintf("Error creating stdout pipe: %v\n", err))
		return 0, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.entrypointCtx.Logger.Error(fmt.Sprintf("Error creating stderr pipe: %v\n", err))
		return 0, err
	}

	err = cmd.Start()
	if err != nil {
		t.entrypointCtx.Logger.Error(fmt.Sprintf("Error starting command: %v\n", err))
		return 0, err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		stdoutBuffer := make([]byte, 64*1024)
		stderrBuffer := make([]byte, 64*1024)

		for {
			stdoutBytesRead, stdoutErr := stdout.Read(stdoutBuffer)
			stderrBytesRead, stderrErr := stderr.Read(stderrBuffer)

			// If either err returned not null and not EOF, break
			if stdoutErr != nil && stdoutErr != io.EOF {
				t.entrypointCtx.Logger.Error(fmt.Sprintf("Error reading from stdout stream: %v\n", stdoutErr))
				break
			}

			if stderrErr != nil && stderrErr != io.EOF {
				t.entrypointCtx.Logger.Error(fmt.Sprintf("Error reading from stderr stream: %v\n", stdoutErr))
				break
			}

			stdoutStr := string(stdoutBuffer[:stdoutBytesRead])
			stderrStr := string(stderrBuffer[:stderrBytesRead])

			if err := streamOutput(stdoutStr, stderrStr); err != nil {
				t.entrypointCtx.Logger.Error("failed to stream output")
				break
			}

			// End of streams
			if stdoutErr != nil && stdoutErr == io.EOF && stderrErr != nil && stderrErr == io.EOF {
				break
			}
		}
	}()

	wg.Wait()

	err = cmd.Wait()

	var exitErr *exec.ExitError
	if err != nil && !errors.As(err, &exitErr) {
		t.entrypointCtx.Logger.Error(fmt.Sprintf("Run finished with error: %v\n", err))
		return 0, err
	}

	exitCode := cmd.ProcessState.ExitCode()
	return uint32(exitCode), nil
}
