package grpc

import (
	"fmt"
	"github.com/spaulg/solo/cli/internal/pkg/solo/orchestrator"
	"github.com/spaulg/solo/cli/internal/pkg/solo/project"
)

type Server struct {
	listener *Listener
}

func NewServer() *Server {
	return &Server{}
}

func (t *Server) Start(
	// todo: refactor to pass hostname & storage directory to decouple
	project *project.Project,
	orchestrator orchestrator.Orchestrator,
) error {
	// Generate certificate files
	certificateGenerator := NewCertificateGenerator(
		orchestrator.GetHostGatewayHostname(),
		project.GetAllServicesStateDirectory(),
	)

	if err := certificateGenerator.Generate(); err != nil {
		return fmt.Errorf("failed to generate grpc server certificate files: %v", err)
	}

	grpcServicePortChannel := make(chan int)
	grpcServiceErrorChannel := make(chan error)

	go func() {
		var err error

		t.listener, err = NewListener(
			certificateGenerator.ServerCertificateFilePath,
			certificateGenerator.ServerPrivateKeyFilePath,
			certificateGenerator.CACertificateFilePath,
		)

		// Start listener and report listening port
		if err != nil {
			grpcServiceErrorChannel <- err

			close(grpcServicePortChannel)
			close(grpcServiceErrorChannel)

			return
		}

		grpcServicePortChannel <- t.listener.Port

		// Start listening
		_ = t.listener.Listen()

		close(grpcServicePortChannel)
		close(grpcServiceErrorChannel)
	}()

	port, ok := <-grpcServicePortChannel
	if !ok {
		return fmt.Errorf("failed to start grpc server: %v", <-grpcServiceErrorChannel)
	}

	grpcServiceLookup := NewGrpcServiceLookup(
		WithHostname(orchestrator.GetHostGatewayHostname()),
		WithPort(port),
		WithClientCertificate(certificateGenerator.ClientCertificateFileName),
		WithClientPrivateKey(certificateGenerator.ClientPrivateKeyFileName),
	)

	grpcServiceLookupFilePath := project.GetAllServicesStateDirectory() + "/grpcservice.yml"
	if err := grpcServiceLookup.MarshallYaml(grpcServiceLookupFilePath); err != nil {
		return fmt.Errorf("failed to generate grpc service lookup definition file: %v", err)
	}

	return nil
}

func (t *Server) Stop() {
	_ = t.listener.Close()
}
