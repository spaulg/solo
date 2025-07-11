package wms

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestWorkflowNameTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowNameTestSuite))
}

type WorkflowNameTestSuite struct {
	suite.Suite
}

func (t *WorkflowNameTestSuite) TestWorkflowNameFromString() {
	t.Equal(FirstPreStart, WorkflowNameFromString("first_pre_start"))
	t.Equal(PreStart, WorkflowNameFromString("pre_start"))
	t.Equal(PostStart, WorkflowNameFromString("post_start"))
	t.Equal(PreStop, WorkflowNameFromString("pre_stop"))
	t.Equal(PostStop, WorkflowNameFromString("post_stop"))
	t.Equal(PreDestroy, WorkflowNameFromString("pre_destroy"))
	t.Equal(PostDestroy, WorkflowNameFromString("post_destroy"))

	t.Equal(Undefined, WorkflowNameFromString("qwerty"))
}

func (t *WorkflowNameTestSuite) TestString() {
	t.Equal("first_pre_start", FirstPreStart.String())
	t.Equal("pre_start", PreStart.String())
	t.Equal("post_start", PostStart.String())
	t.Equal("pre_stop", PreStop.String())
	t.Equal("post_stop", PostStop.String())
	t.Equal("pre_destroy", PreDestroy.String())
	t.Equal("post_destroy", PostDestroy.String())

	t.Equal("unknown", Undefined.String())
}
