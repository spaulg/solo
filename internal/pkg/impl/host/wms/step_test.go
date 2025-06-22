package wms

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestStepTestSuite(t *testing.T) {
	suite.Run(t, new(StepTestSuite))
}
