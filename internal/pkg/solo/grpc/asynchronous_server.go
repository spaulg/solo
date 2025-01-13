package grpc

import (
	"fmt"
	"github.com/spaulg/solo/internal/pkg/common/grpc/services"
	"github.com/spaulg/solo/internal/pkg/solo/grpc/interceptors"
	"github.com/spaulg/solo/internal/pkg/solo/grpc/service_definitions"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"sync/atomic"
)

const hostFileName = "provisioner_host"

type AsynchronousServer struct {
	hostname             string
	port                 uint32
	stateDirectory       string
	transportCredentials credentials.TransportCredentials
	workflowService      *service_definitions.WorkflowServerImpl
	server               *grpc.Server
	grpcServiceErrorCh   chan error
}

func NewAsynchronousServer(
	hostname string,
	port uint16,
	stateDirectory string,
	transportCredentials credentials.TransportCredentials,
	workflowService *service_definitions.WorkflowServerImpl,
) Server {
	return &AsynchronousServer{
		hostname:             hostname,
		port:                 uint32(port),
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

		t.server = grpc.NewServer(
			grpc.Creds(t.transportCredentials),
			grpc.UnaryInterceptor(interceptors.ServiceName),
			grpc.StreamInterceptor(interceptors.ServiceNameStream),
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
		return fmt.Errorf("failed to start grpc service: %v", err)
	}

	return t.writeHostFile()
}

func (t *AsynchronousServer) Stop() {
	t.server.Stop()
}

func (t *AsynchronousServer) writeHostFile() error {
	hostFilePath := path.Join(t.stateDirectory, hostFileName)
	hostFile, err := os.OpenFile(hostFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open host file: %v", err)
	}

	if _, err := hostFile.WriteString(t.hostname + ":" + strconv.Itoa(int(t.port))); err != nil {
		return fmt.Errorf("failed to write to host file: %v", err)
	}

	return nil
}
