package grpc

import (
	"time"

	"github.com/stretchr/testify/mock"
)

type MockContext struct {
	mock.Mock
}

func (m *MockContext) Deadline() (deadline time.Time, ok bool) {
	args := m.Called()
	return args.Get(0).(time.Time), args.Bool(1)
}

func (m *MockContext) Done() <-chan struct{} {
	args := m.Called()
	return args.Get(0).(<-chan struct{})
}

func (m *MockContext) Err() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockContext) Value(key any) any {
	args := m.Called(key)
	return args.Get(0)
}
