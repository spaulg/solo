package interceptors

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

const ServiceNameContextValueName = "ServiceName"

type ServiceName string

type ServiceNameInterceptor struct{}

func NewServiceNameInterceptor() *ServiceNameInterceptor {
	return &ServiceNameInterceptor{}
}

func (t *ServiceNameInterceptor) ServiceNameUnaryInterceptor(
	ctx context.Context,
	req interface{},
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	serviceName, err := findServiceName(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find service name: %v", err)
	}

	ctx = context.WithValue(ctx, ServiceName(ServiceNameContextValueName), serviceName)
	return handler(ctx, req)
}

func (t *ServiceNameInterceptor) ServiceNameStreamInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	_ *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	ctx := ss.Context()
	serviceName, err := findServiceName(ctx)
	if err != nil {
		return fmt.Errorf("failed to find service name: %v", err)
	}

	streamWrapper := NewServerStreamWrapper(ss, context.WithValue(ctx, ServiceName(ServiceNameContextValueName), serviceName))
	return handler(srv, streamWrapper)
}

func findServiceName(ctx context.Context) (string, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "", fmt.Errorf("unable to find peer transport credentials")
	}

	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return "", fmt.Errorf("unable to cast transport credentials to TLSInfo")
	}

	if len(tlsInfo.State.PeerCertificates) == 0 {
		return "", fmt.Errorf("missing peer certificate")
	}

	clientCert := tlsInfo.State.PeerCertificates[0]
	lastIndex := strings.LastIndex(clientCert.Subject.CommonName, ":")

	if lastIndex == -1 {
		return "", fmt.Errorf("invalid subject common name")
	}

	return clientCert.Subject.CommonName[lastIndex+len(":"):], nil
}
