package progress

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestComposeProgressActionTestSuite(t *testing.T) {
	suite.Run(t, new(ComposeProgressActionTestSuite))
}

type ComposeProgressActionTestSuite struct {
	suite.Suite
}

func (t *ComposeProgressActionTestSuite) TestString() {
	t.Equal("Building", Build.String())
	t.Equal("Creating", Create.String())
	t.Equal("Starting", Start.String())
	t.Equal("Stopping", Stop.String())
	t.Equal("Removing", Remove.String())

	t.Equal("Unknown", Unknown.String())
}
