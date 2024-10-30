package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/spaulg/solo/cli/internal/pkg/solo/grpc/service_definitions"
	"github.com/spaulg/solo/shared/pkg/solo/grpc/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
	"os"
	"strconv"
	"strings"
)

type GrpcServer struct {
	listener              net.Listener
	certificateFilePath   string
	certificateKeyPath    string
	caCertificateFilePath string
}

func NewGrpcServer(certificateFilePath string, certificateKeyPath string, caCertificateFilePath string) *GrpcServer {
	return &GrpcServer{
		certificateFilePath:   certificateFilePath,
		certificateKeyPath:    certificateKeyPath,
		caCertificateFilePath: caCertificateFilePath,
	}
}

func (t *GrpcServer) CreateListener() (int, error) {
	// Create listener with randomly assigned port
	// todo: allow a fixed port to be set via config
	listener, err := net.Listen("tcp", "0.0.0.0:12345") // todo: use port 0
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

func (t *GrpcServer) Listen() error {
	tlsConfig, err := t.createTlsConfig()
	if err != nil {
		return err
	}

	tlsCredentials := credentials.NewTLS(tlsConfig)
	server := grpc.NewServer(grpc.Creds(tlsCredentials))
	services.RegisterProvisionerServer(server, &service_definitions.ProvisionerServerImpl{})

	_ = server.Serve(t.listener)
	return nil
}

func (t *GrpcServer) createTlsConfig() (*tls.Config, error) {
	serverCert, err := tls.LoadX509KeyPair(t.certificateFilePath, t.certificateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate: %v", err)
	}

	caCertificate, err := os.ReadFile(t.caCertificateFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %v", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCertificate) {
		return nil, fmt.Errorf("failed to add CA certificate to pool")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    certPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	return tlsConfig, nil
}
