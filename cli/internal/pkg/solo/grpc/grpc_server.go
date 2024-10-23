package grpc

import (
	"fmt"
	"github.com/spaulg/solo/cli/internal/pkg/solo/grpc/service_definitions"
	"github.com/spaulg/solo/shared/pkg/solo/grpc/services"
	"google.golang.org/grpc"
	"net"
	"strconv"
	"strings"
)

type GrpcServer struct {
	listener net.Listener
}

func NewGrpcServer() *GrpcServer {
	return &GrpcServer{}
}

func (t *GrpcServer) CreateListener() (int, error) {
	// Create listener with randomly assigned port
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return 0, err
	}
	t.listener = listener

	// Extract the port from the address
	address := listener.Addr().String()
	lastIndex := strings.LastIndex(address, ":")
	if lastIndex == -1 {
		return 0, fmt.Errorf("unable to find port from address '%s'", address)
	}

	// Return parsed port number
	port, err := strconv.Atoi(address[lastIndex+1:])
	if err != nil {
		return 0, fmt.Errorf("failed to convert address port to integer: %v", err)
	}

	return port, nil
}

func (t *GrpcServer) Listen() {
	server := grpc.NewServer()
	services.RegisterProvisionerServer(server, &service_definitions.ProvisionerServerImpl{})

	_ = server.Serve(t.listener)
}
