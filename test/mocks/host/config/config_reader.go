package config

import (
	config_types "github.com/spaulg/solo/internal/pkg/types/host/config"
	"github.com/stretchr/testify/mock"
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
