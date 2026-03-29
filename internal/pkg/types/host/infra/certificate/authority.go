package certificate

import (
	"crypto/tls"
	"crypto/x509"

	project_types "github.com/spaulg/solo/internal/pkg/types/host/domain/project"
)

type Authority interface {
	GetCACertificate() *tls.Certificate
	ExportCACertificate(project project_types.Project) error
	GenerateCertificate(opts ...CertificateOption) (*tls.Certificate, error)
}

type CertificateOption func(template *x509.Certificate)
