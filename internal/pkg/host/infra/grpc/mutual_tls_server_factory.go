package grpc

import (
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc/credentials"

	"github.com/spaulg/solo/internal/pkg/host/app/context"
	"github.com/spaulg/solo/internal/pkg/host/infra/certificate"
	"github.com/spaulg/solo/internal/pkg/host/infra/grpc/service_definitions"
	events_types "github.com/spaulg/solo/internal/pkg/shared/app/events"
	"github.com/spaulg/solo/internal/pkg/shared/app/wms"
	project_types "github.com/spaulg/solo/internal/pkg/shared/domain"
	container_types "github.com/spaulg/solo/internal/pkg/shared/infra/container"
)

type MutualTLSServerFactory struct {
	soloCtx              *context.CliContext
	eventManager         events_types.Manager
	workflowFactory      wms.WorkflowFactory
	certificateAuthority Authority
}

func NewMutualTLSServerFactory(
	soloCtx *context.CliContext,
	eventManager events_types.Manager,
	workflowFactory wms.WorkflowFactory,
	certificateAuthority Authority,
) *MutualTLSServerFactory {
	return &MutualTLSServerFactory{
		soloCtx:              soloCtx,
		eventManager:         eventManager,
		workflowFactory:      workflowFactory,
		certificateAuthority: certificateAuthority,
	}
}

func (t *MutualTLSServerFactory) Build(
	orchestrator container_types.Orchestrator,
	workflowExecutionTracker wms.WorkflowExecTracker,
	project project_types.Project,
	port int,
) (Server, error) {
	hostname := orchestrator.GetHostGatewayHostname()

	transportCredentials, err := t.buildTransportCredentials(hostname, project)
	if err != nil {
		return nil, err
	}

	workflowService := service_definitions.NewWorkflowService(
		t.soloCtx,
		t.eventManager,
		orchestrator,
		t.workflowFactory,
		workflowExecutionTracker,
	)

	return NewAsynchronousServer(
		orchestrator,
		port,
		project.GetAllServicesStateDirectory(),
		transportCredentials,
		workflowService,
	), nil
}

func (t *MutualTLSServerFactory) buildTransportCredentials(
	hostname string,
	project project_types.Project,
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
		return nil, fmt.Errorf("failed to generate server certificate: %w", err)
	}

	// Generate client certificates for each service
	if err = t.generateClientCertificate(project); err != nil {
		return nil, fmt.Errorf("failed to generate client certificate: %w", err)
	}

	caChain := x509.NewCertPool()
	caChain.AddCert(caCert.Leaf)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*serverCert},
		ClientCAs:    caChain,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS13,
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

func (t *MutualTLSServerFactory) generateClientCertificate(project project_types.Project) error {
	for _, serviceName := range project.Services().ServiceNames() {
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

		privateKey, ok := clientCert.PrivateKey.(*ecdsa.PrivateKey)
		if !ok {
			return fmt.Errorf("private key is not of type *ecdsa.PrivateKey")
		}

		privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
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
