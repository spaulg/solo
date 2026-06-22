package app

import (
	"os"
	"strings"

	"github.com/spaulg/solo/internal/pkg/entrypoint/app/context"
	workflow2 "github.com/spaulg/solo/internal/pkg/entrypoint/app/workflow"
	"github.com/spaulg/solo/internal/pkg/entrypoint/infra/grpc/credentials"
)

const metadataStateFile = "/solo/container/data/metadata_state.yml"

func WorkflowRunnerFactory(entrypointCtx *context.EntrypointContext) (WorkflowRunner, error) {
	targetBytes, err := os.ReadFile("/solo/services_all/provisioner_host")
	if err != nil {
		return nil, err
	}

	target := strings.TrimSpace(string(targetBytes))

	credentialsBuilder, err := credentials.NewMutualTLS()
	if err != nil {
		return nil, err
	}

	metadataState, err := workflow2.LoadMetadataState(metadataStateFile)
	if err != nil {
		return nil, err
	}

	metadataState.Set("hostname", entrypointCtx.InitialHostname)

	return workflow2.NewGrpcWorkflowRunner(entrypointCtx, credentialsBuilder, target, metadataState)
}
