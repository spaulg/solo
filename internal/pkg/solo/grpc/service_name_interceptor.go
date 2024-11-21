package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"strings"
)
import "google.golang.org/grpc"

func ServiceNameInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	serviceName, err := applyServiceName(ctx)
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, "ServiceName", serviceName)
	return handler(ctx, req)
}

/*
func ServiceNameStreamInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	ctx := ss.Context()
	serviceName, err := applyServiceName(ctx)
	if err != nil {
		return err
	}

	wrappedStream := grpc.ServerStreamWrapper{
		ServerStream:   ss,
		WrappedContext: context.WithValue(ctx, "ServiceName", serviceName),
	}

	return handler(srv, wrappedStream)
}
*/

func applyServiceName(ctx context.Context) (string, error) {
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
