package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/spaulg/solo/internal/pkg/solo/certificate"
	"github.com/spaulg/solo/internal/pkg/solo/grpc/service_definitions"
	"github.com/spaulg/solo/internal/pkg/solo/project"
	"google.golang.org/grpc/credentials"
	"os"
)

type MutualTLSServerFactory struct {
	certificateGenerator certificate.CertificateGenerator
	provisionerServer    *service_definitions.ProvisionerServerImpl
}

func NewMutualTLSServerFactory(
	certificateGenerator certificate.CertificateGenerator,
	provisionerServer *service_definitions.ProvisionerServerImpl,
) ServerFactory {
	return &MutualTLSServerFactory{
		certificateGenerator: certificateGenerator,
		provisionerServer:    provisionerServer,
	}
}

func (t *MutualTLSServerFactory) Build(hostname string, port uint16, project *project.Project) (Server, error) {
	transportCredentials, err := t.buildCredentials(project)
	if err != nil {
		return nil, err
	}

	return NewAsynchronousServer(
		hostname,
		port,
		project.GetAllServicesStateDirectory(),
		transportCredentials,
		t.provisionerServer,
	), nil
}

func (t *MutualTLSServerFactory) buildCredentials(project *project.Project) (credentials.TransportCredentials, error) {

	// todo: Make the factory build peer certificates for each service and store in the services state directory

	/*
		tls.Certificate encapsulates both the public and private key. Use this as the value object
		to return from the generation function to decouple persisting the certificate and key to disk

		Extend the persist functionality by decorating the above type with:
		  func (t *tls.Certificate) Export(certPath string, keyPath string) {}



	*/

	var err error
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
