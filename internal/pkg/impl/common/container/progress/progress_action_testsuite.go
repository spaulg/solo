package progress

import "github.com/stretchr/testify/suite"

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
