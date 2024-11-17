package solo

import "github.com/spaulg/solo/agent/internal/pkg/solo/grpc/credentials"

func ProvisionerFactory() (Provisioner, error) {
	credentialsBuilder, err := credentials.NewMutualTLS()
	if err != nil {
		return nil, err
	}

	return NewClient(credentialsBuilder)
}
