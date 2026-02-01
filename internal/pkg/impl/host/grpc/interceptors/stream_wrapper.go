package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ServerStreamWrapper struct {
	wrappedContext context.Context
	wrappedStream  grpc.ServerStream
}

func NewServerStreamWrapper(wrappedContext context.Context, wrappedStream grpc.ServerStream) ServerStreamWrapper {
	return ServerStreamWrapper{
		wrappedContext: wrappedContext,
		wrappedStream:  wrappedStream,
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
