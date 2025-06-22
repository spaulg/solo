package config

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
