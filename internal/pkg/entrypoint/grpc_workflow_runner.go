package entrypoint

import (
	"context"
	"errors"
	"fmt"
	"github.com/spaulg/solo/internal/pkg/common/grpc/services"
	commonworkflow "github.com/spaulg/solo/internal/pkg/common/wms"
	"github.com/spaulg/solo/internal/pkg/entrypoint/grpc/credentials"
	"google.golang.org/grpc"
	"io"
	"strconv"
)

type GrpcWorkflowRunner struct {
	conn           *grpc.ClientConn
	workflowClient services.WorkflowClient
}

type WorkflowStream grpc.BidiStreamingClient[services.WorkflowStreamRequest, services.WorkflowStreamResponse]

func NewGrpcWorkflowRunner(credentialsBuilder credentials.Builder) (WorkflowRunner, error) {
	var err error

	fmt.Println("Connect to grpc server")
	port := 12345                  // todo: obtain from file stored in bind mount
	host := "host.docker.internal" // todo: obtain host from container

	creds, err := credentialsBuilder.Build()
	if err != nil {
		return nil, err
	}

	conn, err := grpc.NewClient(host+":"+strconv.Itoa(port), grpc.WithTransportCredentials(creds))
	if err != nil {
		conn.Close()
		return nil, err
	}

	fmt.Println("Creating new service client")
	client := services.NewWorkflowClient(conn)

	return &GrpcWorkflowRunner{
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
			var exitCode uint32 = 0
			
			fmt.Printf("Running command: %s\n", instruction.RunCommand.Command)

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
