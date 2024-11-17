package entrypoint

import "github.com/spaulg/solo/agent/internal/pkg/entrypoint/grpc/credentials"

func ProvisionerFactory() (Provisioner, error) {
	credentialsBuilder, err := credentials.NewMutualTLS()
	if err != nil {
		return nil, err
	}

	return NewClient(credentialsBuilder)
}
