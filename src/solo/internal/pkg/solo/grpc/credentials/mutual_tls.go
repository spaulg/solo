package credentials

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"google.golang.org/grpc/credentials"
	"os"
)

type MutualTLS struct {
	certificateGenerator CertificateGenerator
}

func NewMutualTLS(
	certificateGenerator CertificateGenerator,
) (Builder, error) {
	return &MutualTLS{
		certificateGenerator: certificateGenerator,
	}, nil
}

func (t *MutualTLS) Build() (credentials.TransportCredentials, error) {
	fmt.Println("Building transport credentials")

	certificatePack, err := t.certificateGenerator.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificates: %v", err)
	}

	serverCert, err := tls.LoadX509KeyPair(certificatePack.ServerCertificateFilePath, certificatePack.ServerPrivateKeyFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate: %v", err)
	}

	caCertificate, err := os.ReadFile(certificatePack.CACertificateFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %v", err)
	}

	caChain := x509.NewCertPool()
	if !caChain.AppendCertsFromPEM(caCertificate) {
		return nil, fmt.Errorf("failed to add CA certificate to pool")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    caChain,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	return credentials.NewTLS(tlsConfig), nil
}
