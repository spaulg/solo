package entrypoint

import (
	"os"
	"strings"

	"github.com/spaulg/solo/internal/pkg/impl/entrypoint/context"
	"github.com/spaulg/solo/internal/pkg/impl/entrypoint/grpc/credentials"
	"github.com/spaulg/solo/internal/pkg/impl/entrypoint/workflow"
	workflow_types "github.com/spaulg/solo/internal/pkg/types/entrypoint/workflow"
)

const metadataStateFile = "/solo/container/data/metadata_state.yml"

func WorkflowRunnerFactory(entrypointCtx *context.EntrypointContext) (workflow_types.WorkflowRunner, error) {
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
