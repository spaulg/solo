package certificate

import (
	"crypto/tls"

	"github.com/stretchr/testify/mock"

	project_types "github.com/spaulg/solo/internal/pkg/types/host/domain"
	certificate_types "github.com/spaulg/solo/internal/pkg/types/host/infra/certificate"
)

type MockAuthority struct {
	mock.Mock
}

func (m *MockAuthority) GetCACertificate() *tls.Certificate {
	args := m.Called()
	return args.Get(0).(*tls.Certificate)
}

func (m *MockAuthority) ExportCACertificate(project project_types.Project) error {
	args := m.Called(project)
	return args.Error(0)
}

func (m *MockAuthority) GenerateCertificate(opts ...certificate_types.CertificateOption) (*tls.Certificate, error) {
	args := m.Called(opts)
	return args.Get(0).(*tls.Certificate), args.Error(1)
}
