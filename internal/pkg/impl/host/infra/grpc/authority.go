package grpc

import (
	"crypto/tls"

	"github.com/spaulg/solo/internal/pkg/impl/host/domain"
	"github.com/spaulg/solo/internal/pkg/impl/host/infra/certificate"
)

type Authority interface {
	GetCACertificate() *tls.Certificate
	ExportCACertificate(project domain.Project) error
	GenerateCertificate(opts ...certificate.Option) (*tls.Certificate, error)
}
