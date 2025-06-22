package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ServerStreamWrapper struct {
	wrappedStream  grpc.ServerStream
	wrappedContext context.Context
}

func NewServerStreamWrapper(wrappedStream grpc.ServerStream, wrappedContext context.Context) ServerStreamWrapper {
	return ServerStreamWrapper{
		wrappedStream:  wrappedStream,
		wrappedContext: wrappedContext,
	}
}

func (t ServerStreamWrapper) SetHeader(metadata metadata.MD) error {
	return t.wrappedStream.SetHeader(metadata)
}

func (t ServerStreamWrapper) SendHeader(metadata metadata.MD) error {
	return t.wrappedStream.SendHeader(metadata)
}

func (t ServerStreamWrapper) SetTrailer(metadata metadata.MD) {
	t.wrappedStream.SetTrailer(metadata)
}

func (t ServerStreamWrapper) Context() context.Context {
	return t.wrappedContext
}

func (t ServerStreamWrapper) SendMsg(message any) error {
	return t.wrappedStream.SendMsg(message)
}

func (t ServerStreamWrapper) RecvMsg(message any) error {
	return t.wrappedStream.RecvMsg(message)
}
