package grpc

import (
	"fmt"
	"github.com/spaulg/solo/internal/pkg/shared/grpc/services"
	"github.com/spaulg/solo/internal/pkg/solo/grpc/credentials"
	"github.com/spaulg/solo/internal/pkg/solo/grpc/service_definitions"
	"google.golang.org/grpc"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
)

type AuthenticatingServer struct {
	hostname           string
	port               uint32
	stateDirectory     string
	credentialsBuilder credentials.Builder
	provisionerServer  *service_definitions.ProvisionerServerImpl
	server             *grpc.Server
	grpcServiceErrorCh chan error
}

func NewServer(
	hostname string,
	port uint16,
	stateDirectory string,
	credentialsBuilder credentials.Builder,
	provisionerServer *service_definitions.ProvisionerServerImpl,
) Server {
	return &AuthenticatingServer{
		hostname:           hostname,
		port:               uint32(port),
		stateDirectory:     stateDirectory,
		credentialsBuilder: credentialsBuilder,
		provisionerServer:  provisionerServer,
		grpcServiceErrorCh: make(chan error, 1),
	}
}

func (t *AuthenticatingServer) Start() error {

	go func() {
		desiredPort := atomic.LoadUint32(&t.port)

		// Create listener
		listener, err := net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(int(desiredPort)))
		if err != nil {
			t.grpcServiceErrorCh <- err
			return
		}

		// Extract the port from the address
		address := listener.Addr().String()
		lastIndex := strings.LastIndex(address, ":")
		if lastIndex == -1 {
			t.grpcServiceErrorCh <- err
			return
		}

		// Discover port number used
		allocatedPort64, err := strconv.ParseUint(address[lastIndex+1:], 10, 32)
		if err != nil {
			t.grpcServiceErrorCh <- err
			return
		}

		grpcCredentials, err := t.credentialsBuilder.Build()
		if err != nil {
			t.grpcServiceErrorCh <- err
			return
		}

		t.server = grpc.NewServer(grpc.Creds(grpcCredentials))
		services.RegisterProvisionerServer(t.server, t.provisionerServer)

		// Report port but only if its different
		allocatedPort := uint32(allocatedPort64)
		if desiredPort != allocatedPort {
			atomic.StoreUint32(&t.port, allocatedPort)
		}

		// Signal main routine the service is about to start successfully
		t.grpcServiceErrorCh <- nil

		if err := t.server.Serve(listener); err != nil {
			t.grpcServiceErrorCh <- err
		}
	}()

	if err := <-t.grpcServiceErrorCh; err != nil {
		return fmt.Errorf("failed to start grpc service: %v", err)
	}

	grpcServiceLookup := NewServiceLookup(t.hostname, uint16(t.port), t.stateDirectory)
	grpcServiceLookup.ApplyCertificatePack(t.credentialsBuilder.GetCertificatePack())

	grpcServiceLookupFilePath := t.stateDirectory + "/grpcservice.yml"
	if err := grpcServiceLookup.MarshallYaml(grpcServiceLookupFilePath); err != nil {
		return fmt.Errorf("failed to generate grpc service lookup definition file: %v", err)
	}

	return nil
}

func (t *AuthenticatingServer) Stop() {
	t.server.Stop()
}
