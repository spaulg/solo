package certificate

import (
	"crypto/x509"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/spaulg/solo/internal/pkg/impl/host/project"
	config_types "github.com/spaulg/solo/internal/pkg/types/host/config"
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
		WithKeyUsage(x509.KeyUsageDigitalSignature),
		WithExtKeyUsage([]x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning, x509.ExtKeyUsageIPSECUser}),
		WithDuration(duration),
		WithCommonName("test.example.com"),
		WithDNSNames([]string{"foo.example.com", "bar.example.com"}),
		WithOrganization([]string{"test Organization"}),
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
	loadedConfig := &config_types.Config{}

	projectFilePath := test.GetTestDataFilePath("certificate/solo.yml")
	loadedProject, err := project.NewProject(projectFilePath, loadedConfig, []string{})
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
