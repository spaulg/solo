package solo

import (
	"fmt"
	"google.golang.org/grpc"
	"net"
)

type GrpcServer struct{}

func NewGrpcServer() *GrpcServer {
	return &GrpcServer{}
}

func (*GrpcServer) Listen() error {
	// todo: use port 0 to get a random port assignment

	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return err
	}

	// todo: share port with parent to give to containers
	// todo: use a data file bind mounted in to the containers instead of env vars
	fmt.Println("address: " + listener.Addr().String())

	server := grpc.NewServer()

	// todo: register protos

	if err := server.Serve(listener); err != nil {
		return err
	}

	return nil
}

// domain modelling to decide structs
// textual analysis to decide this as its quicker to type than to find and use a diagram editor
