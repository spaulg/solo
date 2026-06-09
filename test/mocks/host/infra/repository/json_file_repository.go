package repository

import (
	"github.com/stretchr/testify/mock"
)

type MockJSONFileRepository[T any] struct {
	mock.Mock
}

func (m *MockJSONFileRepository[T]) Save(filePath string, entity T) error {
	args := m.Called(filePath, entity)
	return args.Error(0)
}

func (m *MockJSONFileRepository[T]) Load(filePath string) (T, error) {
	var s T
	var ok bool

	args := m.Called(filePath)

	if s, ok = args.Get(0).(T); ok {
		return s, args.Error(1)
	}

	return s, args.Error(1)
}
