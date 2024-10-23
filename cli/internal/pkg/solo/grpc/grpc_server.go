package grpc

import (
	"fmt"
	"github.com/spaulg/solo/cli/internal/pkg/solo/grpc/service_definitions"
	"github.com/spaulg/solo/shared/pkg/solo/grpc/services"
	"google.golang.org/grpc"
	"net"
)

type GrpcServer struct {
	listener net.Listener
}

func NewGrpcServer() *GrpcServer {
	return &GrpcServer{}
}

func (t *GrpcServer) Start() (int, error) {
	// todo: use port 0 to get a random port assignment

	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return 0, err
	}

	// todo: share port with parent to give to containers
	// todo: use a data file bind mounted in to the containers instead of env vars
	fmt.Println("address: " + listener.Addr().String())
	t.start()

	return 0, nil
}

func (t *GrpcServer) start() error {
	server := grpc.NewServer()
	services.RegisterProvisionerServer(server, &service_definitions.ProvisionerServerImpl{})
	// todo: register protos

	//if err := server.Serve(t.listener); err != nil {
	//	return err
	//}

	return nil
}
