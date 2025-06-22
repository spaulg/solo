package grpc

import (
	"context"

	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
)

type MockServerStream struct {
	mock.Mock
}

func (m *MockServerStream) SetHeader(md metadata.MD) error {
	args := m.Called(md)
	return args.Error(0)
}

func (m *MockServerStream) SendHeader(md metadata.MD) error {
	args := m.Called(md)
	return args.Error(0)
}

func (m *MockServerStream) SetTrailer(md metadata.MD) {
	m.Called(md)
}

func (m *MockServerStream) Context() context.Context {
	args := m.Called()

	if context, ok := args.Get(0).(context.Context); ok {
		return context
	} else {
		return nil
	}
}

func (m *MockServerStream) SendMsg(message any) error {
	args := m.Called(message)
	return args.Error(0)
}

func (m *MockServerStream) RecvMsg(message any) error {
	args := m.Called(message)
	return args.Error(0)
}
