package credentials

import (
	"github.com/spaulg/solo/internal/pkg/solo/certificate"
	"google.golang.org/grpc/credentials"
)

type Builder interface {
	Build() (credentials.TransportCredentials, error)
	GetCertificatePack() *certificate.CertificatePack
}
