package grpc

import (
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/spaulg/solo/internal/pkg/solo/certificate"
	"github.com/spaulg/solo/internal/pkg/solo/container"
	"github.com/spaulg/solo/internal/pkg/solo/grpc/service_definitions"
	"github.com/spaulg/solo/internal/pkg/solo/project"
	"google.golang.org/grpc/credentials"
	"os"
	"time"
)

type MutualTLSServerFactory struct {
	certificateAuthority certificate.Authority
	workflowService      *service_definitions.WorkflowServerImpl
}

func NewMutualTLSServerFactory(
	certificateAuthority certificate.Authority,
	workflowService *service_definitions.WorkflowServerImpl,
) ServerFactory {
	return &MutualTLSServerFactory{
		certificateAuthority: certificateAuthority,
		workflowService:      workflowService,
	}
}

func (t *MutualTLSServerFactory) Build(
	orchestrator container.Orchestrator,
	project *project.Project,
	port uint16,
) (Server, error) {
	hostname := orchestrator.GetHostGatewayHostname()

	transportCredentials, err := t.buildTransportCredentials(hostname, project)
	if err != nil {
		return nil, err
	}

	return NewAsynchronousServer(
		orchestrator,
		port,
		project.GetAllServicesStateDirectory(),
		transportCredentials,
		t.workflowService,
	), nil
}

func (t *MutualTLSServerFactory) buildTransportCredentials(
	hostname string,
	project *project.Project,
) (credentials.TransportCredentials, error) {
	var err error

	err = t.certificateAuthority.ExportCACertificate(project)
	if err != nil {
		return nil, err
	}

	caCert := t.certificateAuthority.GetCACertificate()

	// Generate server certificate
	serverCert, err := t.generateServerCertificate(hostname)
	if err != nil {
		return nil, fmt.Errorf("failed to generate server certificate: %v", err)
	}

	// Generate client certificates for each service
	if err = t.generateClientCertificate(project); err != nil {
		return nil, fmt.Errorf("failed to generate server certificate: %v", err)
	}

	caChain := x509.NewCertPool()
	caChain.AddCert(caCert.Leaf)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*serverCert},
		ClientCAs:    caChain,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	return credentials.NewTLS(tlsConfig), nil
}

func (t *MutualTLSServerFactory) generateServerCertificate(hostname string) (*tls.Certificate, error) {
	return t.certificateAuthority.GenerateCertificate(
		certificate.WithOrganization([]string{"Solo Server"}),
		certificate.WithCommonName(hostname),
		certificate.WithDNSNames([]string{hostname}),
		certificate.WithDuration(3*time.Hour),
		certificate.WithKeyUsage(x509.KeyUsageKeyEncipherment|x509.KeyUsageDigitalSignature),
		certificate.WithExtKeyUsage([]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}),
	)
}

func (t *MutualTLSServerFactory) generateClientCertificate(project *project.Project) error {
	for _, serviceName := range project.ServiceNames() {
		clientCert, err := t.certificateAuthority.GenerateCertificate(
			certificate.WithOrganization([]string{"Solo Client"}),
			certificate.WithCommonName("service:"+serviceName),
			certificate.WithDuration(3*time.Hour),
			certificate.WithKeyUsage(x509.KeyUsageDigitalSignature),
			certificate.WithExtKeyUsage([]x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}),
		)
		if err != nil {
			return err
		}

		stateDirectory := project.GetServiceMountDirectory(serviceName)

		err = os.MkdirAll(stateDirectory, 0700)
		if err != nil {
			return err
		}

		certificateFile, err := os.Create(stateDirectory + "/client.crt")
		if err != nil {
			return err
		}

		defer certificateFile.Close()

		err = pem.Encode(certificateFile, &pem.Block{Type: "CERTIFICATE", Bytes: clientCert.Leaf.Raw})
		if err != nil {
			return err
		}

		keyFile, err := os.Create(stateDirectory + "/client.key")
		if err != nil {
			return err
		}

		defer keyFile.Close()

		privateKeyBytes, err := x509.MarshalECPrivateKey(clientCert.PrivateKey.(*ecdsa.PrivateKey))
		if err != nil {
			return err
		}

		err = pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privateKeyBytes})
		if err != nil {
			return err
		}
	}

	return nil
}
