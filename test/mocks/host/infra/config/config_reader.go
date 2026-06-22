package config

import (
	"github.com/stretchr/testify/mock"

	"github.com/spaulg/solo/internal/pkg/host/domain"
)

type MockConfigReader struct {
	mock.Mock
}

func (t *MockConfigReader) AddConfigPath(path string) error {
	args := t.Called(path)
	return args.Error(0)
}

func (t *MockConfigReader) GetConfig() *domain.Config {
	args := t.Called()
	config := args.Get(0)

	if c, ok := config.(*domain.Config); ok {
		return c
	}

	return nil
}
