package entrypoint

import (
	"github.com/spaulg/solo/internal/pkg/entrypoint/context"
	"github.com/spaulg/solo/internal/pkg/entrypoint/grpc/credentials"
	"github.com/spaulg/solo/internal/pkg/entrypoint/workflow"
	"os"
	"strings"
)

const metadataStateFile = "/solo/container/data/metadata_state.yml"

func WorkflowRunnerFactory(entrypointCtx *context.EntrypointContext) (workflow.WorkflowRunner, error) {
	targetBytes, err := os.ReadFile("/solo/services_all/provisioner_host")
	if err != nil {
		return nil, err
	}

	target := strings.TrimSpace(string(targetBytes))

	credentialsBuilder, err := credentials.NewMutualTLS()
	if err != nil {
		return nil, err
	}

	metadataState, err := workflow.LoadMetadataState(metadataStateFile)
	if err != nil {
		return nil, err
	}

	metadataState.Set("hostname", entrypointCtx.InitialHostname)

	return workflow.NewGrpcWorkflowRunner(entrypointCtx, credentialsBuilder, target, metadataState)
}
