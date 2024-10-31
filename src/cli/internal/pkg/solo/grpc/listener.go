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

type Listener struct {
	listener net.Listener
	server   *grpc.Server

	Port                  uint16
	certificateFilePath   string
	certificateKeyPath    string
	caCertificateFilePath string
}

func NewListener(
	port uint16,
	certificateFilePath string,
	certificateKeyPath string,
	caCertificateFilePath string,
) (*Listener, error) {
	// Create listener with randomly assigned port
	listener, err := net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(int(port)))
	if err != nil {
		return nil, err
	}

	// Extract the port from the address
	address := listener.Addr().String()
	lastIndex := strings.LastIndex(address, ":")
	if lastIndex == -1 {
		return nil, fmt.Errorf("unable to find port from address '%s'", address)
	}

	// Return allocated port number, in events when
	// an automatic port allocation was used
	allocatedPort, err := strconv.ParseUint(address[lastIndex+1:], 10, 16)
	if err != nil {
		return nil, fmt.Errorf("failed to convert address port to integer: %v", err)
	}

	return &Listener{
		listener:              listener,
		Port:                  uint16(allocatedPort),
		certificateFilePath:   certificateFilePath,
		certificateKeyPath:    certificateKeyPath,
		caCertificateFilePath: caCertificateFilePath,
	}, nil
}

func (t *Listener) Listen() error {
	tlsConfig, err := t.createTlsConfig()
	if err != nil {
		return err
	}

	tlsCredentials := credentials.NewTLS(tlsConfig)
	t.server = grpc.NewServer(grpc.Creds(tlsCredentials))
	services.RegisterProvisionerServer(t.server, &service_definitions.ProvisionerServerImpl{})

	_ = t.server.Serve(t.listener)
	return nil
}

func (t *Listener) Close() error {
	t.server.Stop()
	_ = t.listener.Close()

	return nil
}

func (t *Listener) createTlsConfig() (*tls.Config, error) {
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
