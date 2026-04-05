package config

import (
	"github.com/stretchr/testify/mock"

	config_types "github.com/spaulg/solo/internal/pkg/host/domain"
)

type MockConfigReader struct {
	mock.Mock
}

func (t *MockConfigReader) AddConfigPath(path string) error {
	args := t.Called(path)
	return args.Error(0)
}

func (t *MockConfigReader) GetConfig() *config_types.Config {
	args := t.Called()
	config := args.Get(0)

	if c, ok := config.(*config_types.Config); ok {
		return c
	}

	return nil
}
