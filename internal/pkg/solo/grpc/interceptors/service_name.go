package interceptors

import (
	"context"
	"fmt"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"strings"
)
import "google.golang.org/grpc"

const contextValue = "ServiceName"

func ServiceName(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	serviceName, err := findServiceName(ctx)
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, contextValue, serviceName)
	return handler(ctx, req)
}

func ServiceNameStream(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	ctx := ss.Context()
	serviceName, err := findServiceName(ctx)
	if err != nil {
		return err
	}

	streamWrapper := NewServerStreamWrapper(ss, context.WithValue(ctx, contextValue, serviceName))
	return handler(srv, streamWrapper)
}

func findServiceName(ctx context.Context) (string, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "", fmt.Errorf("unexpected peer transport credentials")
	}

	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return "", fmt.Errorf("unexpected peer transport credentials")
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
