package certificate

import (
	"crypto/tls"

	"github.com/stretchr/testify/mock"

	"github.com/spaulg/solo/internal/pkg/host/domain"
	"github.com/spaulg/solo/internal/pkg/host/infra/certificate"
)

type MockAuthority struct {
	mock.Mock
}

func (m *MockAuthority) GetCACertificate() *tls.Certificate {
	args := m.Called()
	return args.Get(0).(*tls.Certificate)
}

func (m *MockAuthority) ExportCACertificate(project domain.Project) error {
	args := m.Called(project)
	return args.Error(0)
}

func (m *MockAuthority) GenerateCertificate(opts ...certificate.Option) (*tls.Certificate, error) {
	args := m.Called(opts)
	return args.Get(0).(*tls.Certificate), args.Error(1)
}
