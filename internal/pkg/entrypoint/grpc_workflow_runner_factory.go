package entrypoint

import "github.com/spaulg/solo/internal/pkg/entrypoint/grpc/credentials"

func WorkflowRunnerFactory() (WorkflowRunner, error) {
	credentialsBuilder, err := credentials.NewMutualTLS()
	if err != nil {
		return nil, err
	}

	return NewGrpcWorkflowRunner(credentialsBuilder)
}
