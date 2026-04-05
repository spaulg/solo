package grpc

import (
	"crypto/tls"

	"github.com/spaulg/solo/internal/pkg/host/infra/certificate"
	project_types "github.com/spaulg/solo/internal/pkg/shared/domain"
)

type Authority interface {
	GetCACertificate() *tls.Certificate
	ExportCACertificate(project project_types.Project) error
	GenerateCertificate(opts ...certificate.CertificateOption) (*tls.Certificate, error)
}
