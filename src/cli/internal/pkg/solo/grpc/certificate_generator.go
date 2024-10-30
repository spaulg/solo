package grpc

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
	"time"
)

type CertificateGenerator struct {
	ServerHostname      string
	CertificateBasePath string

	CACertificateFileName     string
	CAKeyFileName             string
	ServerCertificateFileName string
	ServerPrivateKeyFileName  string
	ClientCertificateFileName string
	ClientPrivateKeyFileName  string

	CACertificateFilePath     string
	CAKeyFilePath             string
	ServerCertificateFilePath string
	ServerPrivateKeyFilePath  string
	ClientCertificateFilePath string
	ClientPrivateKeyFilePath  string

	CACertificate *x509.Certificate
	CAPrivateKey  *ecdsa.PrivateKey
}

func NewCertificateGenerator(
	serverHostname string,
	certificateBasePath string,
) *CertificateGenerator {
	const (
		defaultCACertificateFileName     = "ca.crt"
		defaultCAKeyFileName             = "ca.key"
		defaultServerCertificateFileName = "server.crt"
		defaultServerPrivateKeyFileName  = "server.key"
		defaultClientCertificateFileName = "client.crt"
		defaultClientPrivateKeyFileName  = "client.key"
	)

	t := &CertificateGenerator{
		ServerHostname:            serverHostname,
		CertificateBasePath:       certificateBasePath,
		CACertificateFileName:     defaultCACertificateFileName,
		CAKeyFileName:             defaultCAKeyFileName,
		ServerCertificateFileName: defaultServerCertificateFileName,
		ServerPrivateKeyFileName:  defaultServerPrivateKeyFileName,
		ClientCertificateFileName: defaultClientCertificateFileName,
		ClientPrivateKeyFileName:  defaultClientPrivateKeyFileName,
	}

	// Assign full paths
	t.CACertificateFilePath = t.CertificateBasePath + "/" + t.CACertificateFileName
	t.CAKeyFilePath = t.CertificateBasePath + "/" + t.CAKeyFileName
	t.ServerCertificateFilePath = t.CertificateBasePath + "/" + t.ServerCertificateFileName
	t.ServerPrivateKeyFilePath = t.CertificateBasePath + "/" + t.ServerPrivateKeyFileName
	t.ClientCertificateFilePath = t.CertificateBasePath + "/" + t.ClientCertificateFileName
	t.ClientPrivateKeyFilePath = t.CertificateBasePath + "/" + t.ClientPrivateKeyFileName

	return t
}

func (t *CertificateGenerator) Generate() error {
	_, err := os.Stat(t.CertificateBasePath)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(t.CertificateBasePath, 0750); err != nil {
			return fmt.Errorf("failed to create certificate base path: %v", err)
		}
	}

	if err := t.generateCACertificate(); err != nil {
		return fmt.Errorf("failed to generate CA certificate files: %v", err)
	}

	if err := t.generateServerCertificate(); err != nil {
		return fmt.Errorf("failed to generate server certificate files: %v", err)
	}

	if err := t.generateClientCertificate(); err != nil {
		return fmt.Errorf("failed to generate client certificate files: %v", err)
	}

	return nil
}

func (t *CertificateGenerator) generateCACertificate() error {
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

	certificate, key, err := t.generateCertificate(&caTemplate, t.CACertificateFilePath, t.CAKeyFilePath, nil, nil)
	if err != nil {
		return err
	}

	t.CACertificate = certificate
	t.CAPrivateKey = key

	return nil
}

func (t *CertificateGenerator) generateServerCertificate() error {
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

		DNSNames: []string{t.ServerHostname},
	}

	_, _, err := t.generateCertificate(
		&certificateTemplate,
		t.ServerCertificateFilePath,
		t.ServerPrivateKeyFilePath,
		t.CACertificate,
		t.CAPrivateKey,
	)

	if err != nil {
		return err
	}

	return nil
}

func (t *CertificateGenerator) generateClientCertificate() error {
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
		t.ClientCertificateFilePath,
		t.ClientPrivateKeyFilePath,
		t.CACertificate,
		t.CAPrivateKey,
	)

	if err != nil {
		return err
	}

	return nil
}

func (t *CertificateGenerator) generateCertificate(
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
