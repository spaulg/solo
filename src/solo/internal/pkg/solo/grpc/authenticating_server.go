package grpc

import (
	"github.com/spaulg/solo/cli/internal/pkg/solo/grpc/credentials"
	"github.com/spaulg/solo/cli/internal/pkg/solo/grpc/service_definitions"
	"github.com/spaulg/solo/shared/pkg/solo/grpc/services"
	"google.golang.org/grpc"
	"net"
	"strconv"
	"strings"
)

type AuthenticatingServer struct {
	hostname           string
	port               uint16
	stateDirectory     string
	credentialsBuilder credentials.Builder
	server             *grpc.Server
}

func NewServer(
	hostname string,
	port uint16,
	stateDirectory string,
	credentialsBuilder credentials.Builder,
) Server {
	return &AuthenticatingServer{
		hostname:           hostname,
		port:               port,
		stateDirectory:     stateDirectory,
		credentialsBuilder: credentialsBuilder,
	}
}

func (t *AuthenticatingServer) Start() error {
	grpcServicePortChannel := make(chan uint16)
	grpcServiceErrorChannel := make(chan error)

	go func() {
		var err error

		// Create listener with randomly assigned port
		listener, err := net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(int(t.port)))
		if err != nil {
			return
			//nil, err
		}

		// Extract the port from the address
		address := listener.Addr().String()
		lastIndex := strings.LastIndex(address, ":")
		if lastIndex == -1 {
			return
			//nil, fmt.Errorf("unable to find port from address '%s'", address)
		}

		// Return allocated port number, in events when
		// an automatic port allocation was used
		_, err = strconv.ParseUint(address[lastIndex+1:], 10, 16)
		if err != nil {
			return
			//nil, fmt.Errorf("failed to convert address port to integer: %v", err)
		}

		// todo: broken
		//grpcServicePortChannel <- uint16(allocatedPort)

		builder, err := t.credentialsBuilder.Build()
		if err != nil {
			return
		}

		t.server = grpc.NewServer(grpc.Creds(builder))
		services.RegisterProvisionerServer(t.server, &service_definitions.ProvisionerServerImpl{})

		_ = t.server.Serve(listener)

		close(grpcServicePortChannel)
		close(grpcServiceErrorChannel)
	}()

	//port, ok := <-grpcServicePortChannel
	//if !ok {
	//	return fmt.Errorf("failed to start grpc server: %v", <-grpcServiceErrorChannel)
	//}
	//
	//grpcServiceLookup := NewServiceLookup(
	//	// todo: refactor these out - they're all required so shouldn't be configured using optional with pattern
	//	WithHostname(t.hostname),
	//	WithPort(port),
	//	WithClientCertificate(t.certificateGenerator.ClientCertificateFileName),
	//	WithClientPrivateKey(t.certificateGenerator.ClientPrivateKeyFileName),
	//)
	//
	//grpcServiceLookupFilePath := t.stateDirectory + "/grpcservice.yml"
	//if err := grpcServiceLookup.MarshallYaml(grpcServiceLookupFilePath); err != nil {
	//	return fmt.Errorf("failed to generate grpc service lookup definition file: %v", err)
	//}

	return nil
}

func (t *AuthenticatingServer) Stop() {
	//_ = t.listener.Close()
}
