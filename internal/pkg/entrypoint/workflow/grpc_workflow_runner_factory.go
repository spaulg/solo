package workflow

import (
	"github.com/spaulg/solo/internal/pkg/entrypoint/context"
	"github.com/spaulg/solo/internal/pkg/entrypoint/grpc/credentials"
)

func WorkflowRunnerFactory(entrypointCtx *context.EntrypointContext) (WorkflowRunner, error) {
	credentialsBuilder, err := credentials.NewMutualTLS()
	if err != nil {
		return nil, err
	}

	return NewGrpcWorkflowRunner(entrypointCtx, credentialsBuilder)
}
