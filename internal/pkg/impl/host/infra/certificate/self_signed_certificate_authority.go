package certificate

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"

	project_types "github.com/spaulg/solo/internal/pkg/types/host/domain/project"
	certificate_types "github.com/spaulg/solo/internal/pkg/types/host/infra/certificate"
)

type SelfSignedCertificateAuthority struct {
	caCertificate *tls.Certificate
}

func WithDuration(duration time.Duration) certificate_types.CertificateOption {
	return func(template *x509.Certificate) {
		template.NotAfter = time.Now().Add(duration)
	}
}

func WithKeyUsage(keyUsage x509.KeyUsage) certificate_types.CertificateOption {
	return func(template *x509.Certificate) {
		template.KeyUsage = keyUsage
	}
}

func WithExtKeyUsage(extKeyUsage []x509.ExtKeyUsage) certificate_types.CertificateOption {
	return func(template *x509.Certificate) {
		template.ExtKeyUsage = extKeyUsage
	}
}

func WithOrganization(organization []string) certificate_types.CertificateOption {
	return func(template *x509.Certificate) {
		template.Subject.Organization = organization
	}
}

func WithCommonName(commonName string) certificate_types.CertificateOption {
	return func(template *x509.Certificate) {
		template.Subject.CommonName = commonName
	}
}

func WithDNSNames(dnsNames []string) certificate_types.CertificateOption {
	return func(template *x509.Certificate) {
		template.DNSNames = dnsNames
	}
}

func NewCertificateAuthority() (certificate_types.Authority, error) {
	certificateTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Solo CA"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 year
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	clientCertDER, err := x509.CreateCertificate(
		rand.Reader,
		&certificateTemplate,
		&certificateTemplate,
		&privateKey.PublicKey,
		privateKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key bytes: %w", err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privateKeyBytes})
	if keyPEM == nil {
		return nil, fmt.Errorf("failed to encode private key")
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientCertDER})
	if certPEM == nil {
		return nil, fmt.Errorf("failed to encode client certificate")
	}

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to load keypair: %w", err)
	}

	cert.Leaf, err = x509.ParseCertificate(clientCertDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return &SelfSignedCertificateAuthority{
		caCertificate: &cert,
	}, nil
}

func (t *SelfSignedCertificateAuthority) GetCACertificate() *tls.Certificate {
	return t.caCertificate
}

func (t *SelfSignedCertificateAuthority) ExportCACertificate(project project_types.Project) error {
	stateDirectory := project.GetAllServicesStateDirectory()

	err := os.MkdirAll(stateDirectory, 0700)
	if err != nil {
		return err
	}

	certFile, err := os.Create(stateDirectory + "/ca.crt")
	if err != nil {
		return err
	}

	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: t.caCertificate.Certificate[0]})
	if err != nil {
		return err
	}

	return nil
}

func (t *SelfSignedCertificateAuthority) GenerateCertificate(
	opts ...certificate_types.CertificateOption,
) (*tls.Certificate, error) {
	certificateTemplate := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(3 * time.Hour),
		BasicConstraintsValid: true,
	}

	for _, opt := range opts {
		opt(&certificateTemplate)
	}

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	certDER, err := x509.CreateCertificate(
		rand.Reader,
		&certificateTemplate,
		t.caCertificate.Leaf,
		&privateKey.PublicKey,
		t.caCertificate.PrivateKey,
	)
	if err != nil {
		return nil, err
	}

	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privateKeyBytes})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	cert.Leaf, err = x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return &cert, nil
}
