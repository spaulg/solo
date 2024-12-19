package certificate

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/spaulg/solo/internal/pkg/solo/project"
)

type Authority interface {
	GetCACertificate() *tls.Certificate
	ExportCACertificate(project *project.Project) error
	GenerateCertificate(opts ...CertificateOption) (*tls.Certificate, error)
}

type CertificateOption func(template *x509.Certificate)
