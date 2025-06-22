package grpc

import "github.com/stretchr/testify/mock"

type MockAsynchronousServer struct {
	mock.Mock
}

func (m *MockAsynchronousServer) Start() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAsynchronousServer) Stop() {
	m.Called()
}
