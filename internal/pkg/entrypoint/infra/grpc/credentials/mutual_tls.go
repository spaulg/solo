package credentials

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"

	"google.golang.org/grpc/credentials"
)

type MutualTLS struct{}

func NewMutualTLS() (*MutualTLS, error) {
	return &MutualTLS{}, nil
}

func (t *MutualTLS) Build() (credentials.TransportCredentials, error) {
	clientCert, err := tls.LoadX509KeyPair("/solo/service/data/client.crt", "/solo/service/data/client.key")
	if err != nil {
		log.Fatalf("failed to load client certificate: %v", err)
		return nil, err
	}

	// Load the CA certificate
	caCert, err := os.ReadFile("/solo/services_all/ca.crt")
	if err != nil {
		log.Fatalf("failed to read CA certificate: %v", err)
		return nil, err
	}

	// Create a cert pool and add the CA certificate
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		log.Fatalf("failed to add CA certificate to pool")
		return nil, err
	}

	// Create a TLS config with the client certificate and CA
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
		MinVersion:   tls.VersionTLS13,
	}

	// Create gRPC credentialss
	creds := credentials.NewTLS(tlsConfig)
	return creds, nil
}
