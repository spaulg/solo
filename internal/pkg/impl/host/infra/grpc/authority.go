package grpc

import (
	"crypto/tls"

	"github.com/spaulg/solo/internal/pkg/impl/host/infra/certificate"
	project_types "github.com/spaulg/solo/internal/pkg/types/host/domain"
)

type Authority interface {
	GetCACertificate() *tls.Certificate
	ExportCACertificate(project project_types.Project) error
	GenerateCertificate(opts ...certificate.Option) (*tls.Certificate, error)
}
