package workflow

import (
	"context"
	"errors"
	"fmt"
	"github.com/spaulg/solo/internal/pkg/common/grpc/services"
	commonworkflow "github.com/spaulg/solo/internal/pkg/common/wms"
	entrypointcontext "github.com/spaulg/solo/internal/pkg/entrypoint/context"
	"github.com/spaulg/solo/internal/pkg/entrypoint/grpc/credentials"
	"google.golang.org/grpc"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type GrpcWorkflowRunner struct {
	entrypointCtx  *entrypointcontext.EntrypointContext
	conn           *grpc.ClientConn
	workflowClient services.WorkflowClient
}

type WorkflowStream grpc.BidiStreamingClient[services.WorkflowStreamRequest, services.WorkflowStreamResponse]

func NewGrpcWorkflowRunner(
	entrypointCtx *entrypointcontext.EntrypointContext,
	credentialsBuilder credentials.Builder,
) (WorkflowRunner, error) {
	entrypointCtx.Logger.Info("Connect to grpc server")

	creds, err := credentialsBuilder.Build()
	if err != nil {
		return nil, err
	}

	targetBytes, err := os.ReadFile("/solo/services_all/provisioner_host")
	if err != nil {
		return nil, err
	}

	target := strings.TrimSpace(string(targetBytes))
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(creds))
	if err != nil {
		conn.Close()
		return nil, err
	}

	entrypointCtx.Logger.Info("Creating new service client")
	client := services.NewWorkflowClient(conn)

	return &GrpcWorkflowRunner{
		entrypointCtx:  entrypointCtx,
		conn:           conn,
		workflowClient: client,
	}, nil
}

func (t *GrpcWorkflowRunner) Execute(workflowName commonworkflow.Name) {
	stream, err := t.buildStream(workflowName)
	if err != nil {
		panic(err)
	}

	defer stream.CloseSend()

	for {
		instruction, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
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

					return nil
				},
			)

			if err != nil {
				panic(err)
			}

			if err := stream.Send(&services.WorkflowStreamRequest{
				Result: services.WorkflowResult_RUN_COMMAND_RESULT,
				RunCommandResult: &services.WorkflowRunResult{
					ExitCode: &exitCode,
				},
			}); err != nil {
				panic(err)
			}

		case services.WorkflowAction_COMPLETE_ACTION:
			break

		default:
			panic(err)
		}
	}
}

func (t *GrpcWorkflowRunner) Close() error {
	return t.conn.Close()
}

func (t *GrpcWorkflowRunner) buildStream(workflowName commonworkflow.Name) (WorkflowStream, error) {
	switch workflowName {
	case commonworkflow.Build:
		return t.workflowClient.BuildWorkflowStream(context.Background())
	case commonworkflow.PreStart:
		return t.workflowClient.PreStartWorkflowStream(context.Background())
	case commonworkflow.PostStart:
		return t.workflowClient.PostStartWorkflowStream(context.Background())
	default:
		return nil, errors.New("invalid wms name")
	}
}

func (t *GrpcWorkflowRunner) execute(
	command string,
	arguments []string,
	workingDirectory *string,
	streamOutput func(stdout string, stderr string) error,
) (uint32, error) {
	cmd := exec.Command(command, arguments...)

	if workingDirectory == nil {
		cmd.Dir = "/"
	} else {
		cmd.Dir = *workingDirectory
	}

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
			if (stdoutErr != nil && stdoutErr != io.EOF) || (stderrErr != nil && stderrErr != io.EOF) {
				t.entrypointCtx.Logger.Error(fmt.Sprintf("Error reading from output stream: %v\n", stdoutErr))
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

	err = cmd.Wait()
	exitCode := cmd.ProcessState.ExitCode()

	var exitErr *exec.ExitError
	if err != nil && !errors.As(err, &exitErr) || exitCode == -1 {
		t.entrypointCtx.Logger.Error(fmt.Sprintf("Command finished with error: %v\n", err))
		return 0, nil
	}

	wg.Wait()

	return uint32(exitCode), nil
}
