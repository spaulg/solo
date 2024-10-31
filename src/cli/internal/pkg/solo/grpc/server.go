package grpc

import (
	"fmt"
)

type Server struct {
	listener       *Listener
	hostname       string
	port           uint16
	stateDirectory string
}

func NewServer(hostname string, port uint16, stateDirectory string) *Server {
	return &Server{
		hostname:       hostname,
		port:           port,
		stateDirectory: stateDirectory,
	}
}

func (t *Server) Start() error {
	// Generate certificate files
	certificateGenerator := NewCertificateGenerator(t.hostname, t.stateDirectory)
	if err := certificateGenerator.Generate(); err != nil {
		return fmt.Errorf("failed to generate grpc server certificate files: %v", err)
	}

	grpcServicePortChannel := make(chan uint16)
	grpcServiceErrorChannel := make(chan error)

	go func() {
		var err error

		t.listener, err = NewListener(
			t.port,
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

	grpcServiceLookup := NewServiceLookup(
		// todo: refactor these out - they're all required so shouldn't be configured using optional with pattern
		WithHostname(t.hostname),
		WithPort(port),
		WithClientCertificate(certificateGenerator.ClientCertificateFileName),
		WithClientPrivateKey(certificateGenerator.ClientPrivateKeyFileName),
	)

	grpcServiceLookupFilePath := t.stateDirectory + "/grpcservice.yml"
	if err := grpcServiceLookup.MarshallYaml(grpcServiceLookupFilePath); err != nil {
		return fmt.Errorf("failed to generate grpc service lookup definition file: %v", err)
	}

	return nil
}

func (t *Server) Stop() {
	_ = t.listener.Close()
}
