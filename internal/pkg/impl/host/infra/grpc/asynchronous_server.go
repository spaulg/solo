package grpc

import (
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"sync/atomic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/spaulg/solo/internal/pkg/impl/common/infra/grpc/services"
	"github.com/spaulg/solo/internal/pkg/impl/host/infra/grpc/interceptors"
	container_types "github.com/spaulg/solo/internal/pkg/types/host/infra/container"
	grpc_types "github.com/spaulg/solo/internal/pkg/types/host/infra/grpc"
)

const hostFileName = "provisioner_host"

type AsynchronousServer struct {
	orchestrator         container_types.Orchestrator
	port                 uint32
	stateDirectory       string
	transportCredentials credentials.TransportCredentials
	workflowService      services.WorkflowServer
	server               *grpc.Server
	grpcServiceErrorCh   chan error
}

func NewAsynchronousServer(
	orchestrator container_types.Orchestrator,
	port int,
	stateDirectory string,
	transportCredentials credentials.TransportCredentials,
	workflowService services.WorkflowServer,
) grpc_types.Server {
	return &AsynchronousServer{
		orchestrator:         orchestrator,
		port:                 uint32(port), // nolint:gosec
		stateDirectory:       stateDirectory,
		transportCredentials: transportCredentials,
		workflowService:      workflowService,
		grpcServiceErrorCh:   make(chan error, 1),
	}
}

func (t *AsynchronousServer) Start() error {

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

		serviceNameInterceptor := interceptors.NewServiceNameInterceptor()
		containerNameInterceptor := interceptors.NewContainerNameInterceptor(t.orchestrator)
		firstPreStartCompleteInterceptor := interceptors.NewFirstContainerCompleteInterceptor(t.orchestrator)

		t.server = grpc.NewServer(
			grpc.Creds(t.transportCredentials),
			grpc.ChainUnaryInterceptor(
				serviceNameInterceptor.ServiceNameUnaryInterceptor,
				firstPreStartCompleteInterceptor.FirstContainerCompleteUnaryInterceptor,
				containerNameInterceptor.ContainerNameUnaryInterceptor,
			),
			grpc.ChainStreamInterceptor(
				serviceNameInterceptor.ServiceNameStreamInterceptor,
				firstPreStartCompleteInterceptor.FirstContainerCompleteStreamInterceptor,
				containerNameInterceptor.ContainerNameStreamInterceptor,
			),
		)

		services.RegisterWorkflowServer(t.server, t.workflowService)

		// Report port but only if it's different
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
		return fmt.Errorf("failed to start grpc service: %w", err)
	}

	return t.writeHostFile()
}

func (t *AsynchronousServer) Stop() {
	t.server.Stop()
}

func (t *AsynchronousServer) writeHostFile() error {
	hostname := t.orchestrator.GetHostGatewayHostname()
	hostFilePath := path.Join(t.stateDirectory, hostFileName)
	hostFile, err := os.OpenFile(hostFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open host file: %w", err)
	}

	if _, err := hostFile.WriteString(hostname + ":" + strconv.Itoa(int(t.port))); err != nil {
		return fmt.Errorf("failed to write to host file: %w", err)
	}

	return nil
}
