package self_signed

import (
	"crypto/x509"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/spaulg/solo/internal/pkg/host/domain"
	"github.com/spaulg/solo/internal/pkg/host/domain/compose"
	"github.com/spaulg/solo/internal/pkg/host/infra/certificate"
	"github.com/spaulg/solo/test"
)

func TestTestCertificateAuthority(t *testing.T) {
	suite.Run(t, new(TestCertificateAuthority))
}

type TestCertificateAuthority struct {
	suite.Suite
}

func (t *TestCertificateAuthority) TestNewCertificateAuthority() {
	ca, err := NewCertificateAuthority()
	t.NoError(err)

	caCert := ca.GetCACertificate()
	t.NotNil(caCert)
}

func (t *TestCertificateAuthority) TestGenerateCertificate() {
	ca, err := NewCertificateAuthority()
	t.NoError(err)
	duration := 24 * time.Hour

	cert, err := ca.GenerateCertificate(
		certificate.WithKeyUsage(x509.KeyUsageDigitalSignature),
		certificate.WithExtKeyUsage([]x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning, x509.ExtKeyUsageIPSECUser}),
		certificate.WithDuration(duration),
		certificate.WithCommonName("test.example.com"),
		certificate.WithDNSNames([]string{"foo.example.com", "bar.example.com"}),
		certificate.WithOrganization([]string{"test Organization"}),
	)
	t.NoError(err)

	t.Equal(x509.KeyUsageDigitalSignature, cert.Leaf.KeyUsage)
	t.Equal([]x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning, x509.ExtKeyUsageIPSECUser}, cert.Leaf.ExtKeyUsage)

	t.Equal(time.Now().Add(duration).Truncate(time.Second).UTC().String(), cert.Leaf.NotAfter.String())

	t.Equal("test.example.com", cert.Leaf.Subject.CommonName)
	t.Equal([]string{"foo.example.com", "bar.example.com"}, cert.Leaf.DNSNames)

	t.Equal([]string{"test Organization"}, cert.Leaf.Subject.Organization)
}

func (t *TestCertificateAuthority) TestExportCACertificate() {
	loadedConfig := &domain.Config{}

	projectFilePath := test.GetTestDataFilePath("certificate/solo.yml")
	loadedProject, err := compose.NewProject(projectFilePath, loadedConfig, []string{})
	t.NoError(err)

	ca, err := NewCertificateAuthority()
	t.NoError(err)

	err = ca.ExportCACertificate(loadedProject)
	t.NoError(err)

	stateDirectory := loadedProject.GetAllServicesStateDirectory()
	certPath := stateDirectory + "/ca.crt"
	t.FileExists(certPath)

	// Clean up
	os.Remove(certPath)
	os.Remove(stateDirectory)
}
