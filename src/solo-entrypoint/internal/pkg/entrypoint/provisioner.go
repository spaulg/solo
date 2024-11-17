package entrypoint

import (
	"context"
	"fmt"
	"github.com/spaulg/solo/agent/internal/pkg/entrypoint/grpc/credentials"
	"github.com/spaulg/solo/shared/pkg/solo/grpc/services"
	"google.golang.org/grpc"
	"io"
	"strconv"
)

type Provisioner interface {
	io.Closer
	Finish()
}

type GrpcProvisioner struct {
	conn              *grpc.ClientConn
	provisionerClient services.ProvisionerClient
}

func NewClient(credentialsBuilder credentials.Builder) (Provisioner, error) {
	var err error

	fmt.Println("Connect to grpc server")
	port := 12345                  // todo: obtain from file stored in bind mount
	host := "host.docker.internal" // todo: obtain host from orchestrator

	creds, err := credentialsBuilder.Build()
	if err != nil {
		return nil, err
	}

	conn, err := grpc.NewClient(host+":"+strconv.Itoa(port), grpc.WithTransportCredentials(creds))
	if err != nil {
		conn.Close()
		return nil, err
	}

	fmt.Println("Creating new service client")
	client := services.NewProvisionerClient(conn)

	return &GrpcProvisioner{
		conn:              conn,
		provisionerClient: client,
	}, nil
}

func (t *GrpcProvisioner) Finish() {
	// Signal provisioning run as finished
	fmt.Println("Calling finish")
	_, err := t.provisionerClient.NotifyProvisionerComplete(context.Background(), &services.NotifyProvisionerCompleteRequest{})
	if err != nil {
		panic(err)
	}
}

func (t *GrpcProvisioner) Close() error {
	return t.conn.Close()
}
