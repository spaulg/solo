package credentials

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"
)

type SelfSignedCertificateGenerator struct {
	serverHostname      string
	certificateBasePath string

	caCertificateFilePath     string
	caKeyFilePath             string
	serverCertificateFilePath string
	serverPrivateKeyFilePath  string
	clientCertificateFilePath string
	clientPrivateKeyFilePath  string

	caCertificate *x509.Certificate
	caPrivateKey  *ecdsa.PrivateKey
}

func NewCertificateGenerator(
	serverHostname string,
	certificateBasePath string,
) (CertificateGenerator, error) {
	const (
		defaultCACertificateFileName     = "ca.crt"
		defaultCAKeyFileName             = "ca.key"
		defaultServerCertificateFileName = "server.crt"
		defaultServerPrivateKeyFileName  = "server.key"
		defaultClientCertificateFileName = "client.crt"
		defaultClientPrivateKeyFileName  = "client.key"
	)

	serverHostname = strings.TrimSpace(serverHostname)
	if len(serverHostname) == 0 {
		return nil, InvalidHostname
	}

	certificateBasePath = strings.TrimSpace(certificateBasePath)
	if len(certificateBasePath) == 0 {
		return nil, InvalidCertificateBasePath
	}

	t := &SelfSignedCertificateGenerator{
		serverHostname:      serverHostname,
		certificateBasePath: certificateBasePath,
	}

	// Assign full paths
	t.caCertificateFilePath = t.certificateBasePath + "/" + defaultCACertificateFileName
	t.caKeyFilePath = t.certificateBasePath + "/" + defaultCAKeyFileName
	t.serverCertificateFilePath = t.certificateBasePath + "/" + defaultServerCertificateFileName
	t.serverPrivateKeyFilePath = t.certificateBasePath + "/" + defaultServerPrivateKeyFileName
	t.clientCertificateFilePath = t.certificateBasePath + "/" + defaultClientCertificateFileName
	t.clientPrivateKeyFilePath = t.certificateBasePath + "/" + defaultClientPrivateKeyFileName

	return t, nil
}

func (t *SelfSignedCertificateGenerator) Generate() (*CertificatePack, error) {
	_, err := os.Stat(t.certificateBasePath)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(t.certificateBasePath, 0750); err != nil {
			return nil, fmt.Errorf("failed to create certificate base path: %v", err)
		}
	}

	var certificatePack = NewCertificatePack()

	if err := t.generateCACertificate(certificatePack); err != nil {
		return nil, fmt.Errorf("failed to generate CA certificate files: %v", err)
	}

	if err := t.generateServerCertificate(certificatePack); err != nil {
		return nil, fmt.Errorf("failed to generate server certificate files: %v", err)
	}

	if err := t.generateClientCertificate(certificatePack); err != nil {
		return nil, fmt.Errorf("failed to generate client certificate files: %v", err)
	}

	return certificatePack, nil
}

func (t *SelfSignedCertificateGenerator) generateCACertificate(certificatePack *CertificatePack) error {
	caTemplate := x509.Certificate{
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

	certificate, key, err := t.generateCertificate(&caTemplate, t.caCertificateFilePath, t.caKeyFilePath, nil, nil)
	if err != nil {
		return err
	}

	certificatePack.CACertificateFilePath = t.caCertificateFilePath
	certificatePack.CAKeyFilePath = t.caKeyFilePath

	t.caCertificate = certificate
	t.caPrivateKey = key

	return nil
}

func (t *SelfSignedCertificateGenerator) generateServerCertificate(certificatePack *CertificatePack) error {
	certificateTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Solo server"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(3 * time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,

		DNSNames: []string{t.serverHostname},
	}

	_, _, err := t.generateCertificate(
		&certificateTemplate,
		t.serverCertificateFilePath,
		t.serverPrivateKeyFilePath,
		t.caCertificate,
		t.caPrivateKey,
	)

	certificatePack.ServerCertificateFilePath = t.serverCertificateFilePath
	certificatePack.ServerPrivateKeyFilePath = t.serverPrivateKeyFilePath

	if err != nil {
		return err
	}

	return nil
}

func (t *SelfSignedCertificateGenerator) generateClientCertificate(certificatePack *CertificatePack) error {
	clientTemplate := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization: []string{"Solo Client"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour), // 1 year

		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	_, _, err := t.generateCertificate(
		&clientTemplate,
		t.clientCertificateFilePath,
		t.clientPrivateKeyFilePath,
		t.caCertificate,
		t.caPrivateKey,
	)

	certificatePack.ClientCertificateFilePath = t.clientCertificateFilePath
	certificatePack.ClientPrivateKeyFilePath = t.clientPrivateKeyFilePath

	if err != nil {
		return err
	}

	return nil
}

func (t *SelfSignedCertificateGenerator) generateCertificate(
	certificateTemplate *x509.Certificate,
	certificateFileName string,
	privateKeyFileName string,
	parentCertificate *x509.Certificate,
	parentPrivateKey *ecdsa.PrivateKey,
) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	if parentCertificate == nil {
		parentCertificate = certificateTemplate
	}

	if parentPrivateKey == nil {
		parentPrivateKey = privateKey
	}

	clientCertDER, err := x509.CreateCertificate(
		rand.Reader,
		certificateTemplate,
		parentCertificate,
		&privateKey.PublicKey,
		parentPrivateKey,
	)

	if err != nil {
		return nil, nil, err
	}

	// Save certificate to file
	certOut, err := os.Create(certificateFileName)
	if err != nil {
		return nil, nil, err
	}

	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: clientCertDER}); err != nil {
		return nil, nil, err
	}

	// Save private key to file
	keyOut, err := os.Create(privateKeyFileName)
	if err != nil {
		return nil, nil, err
	}

	defer keyOut.Close()

	privBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, nil, err
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes}); err != nil {
		return nil, nil, err
	}

	caCert, err := x509.ParseCertificate(clientCertDER)
	if err != nil {
		return nil, nil, err
	}

	return caCert, privateKey, nil
}
