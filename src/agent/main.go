package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/spaulg/solo/shared/pkg/solo/grpc/services"
)

func main() {
	var err error

	clientCert, err := tls.LoadX509KeyPair("/solo/services_all/client.crt", "/solo/services_all/client.key")
	if err != nil {
		log.Fatalf("failed to load client certificate: %v", err)
	}

	// Load the CA certificate
	caCert, err := os.ReadFile("/solo/services_all/ca.crt")
	if err != nil {
		log.Fatalf("failed to read CA certificate: %v", err)
	}

	// Create a cert pool and add the CA certificate
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		log.Fatalf("failed to add CA certificate to pool")
	}

	// Create a TLS config with the client certificate and CA
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}

	// Create gRPC credentialss
	creds := credentials.NewTLS(tlsConfig)

	fmt.Println("Connect to grpc server")
	port := 12345 // todo: obtain from file stored in bind mount
	// todo: obtain host from orchestrator
	// todo: implement mtls for auth
	conn, err := grpc.NewClient("host.docker.internal:"+strconv.Itoa(port), grpc.WithTransportCredentials(creds))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("Creating new service client")
	client := services.NewProvisionerClient(conn)

	// Signal provisioning run as finished
	fmt.Println("Calling finish")
	_, err = client.NotifyProvisionerComplete(context.Background(), &services.NotifyProvisionerCompleteRequest{})
	if err != nil {
		panic(err)
	}

	if strings.HasPrefix(os.Args[1], "/") {
		// Full path of executable given
		err = syscall.Exec(os.Args[1], os.Args[1:], nil)
	} else {
		// todo: Requires $PATH env var and needs a shell
		args := []string{"/bin/sh", "-c"}
		args = append(args, strings.Join(os.Args[1:], " "))

		err = syscall.Exec("/bin/sh", args, nil)
	}

	panic(err)
}
