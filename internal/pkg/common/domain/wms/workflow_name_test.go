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
	// Start
	t.Equal(FirstPreStartContainer, WorkflowNameFromString("first_pre_start_container"))
	t.Equal(FirstPreStartService, WorkflowNameFromString("first_pre_start_service"))
	t.Equal(PreStartContainer, WorkflowNameFromString("pre_start_container"))
	t.Equal(PreStartService, WorkflowNameFromString("pre_start_service"))
	t.Equal(PostStartContainer, WorkflowNameFromString("post_start_container"))
	t.Equal(PostStartService, WorkflowNameFromString("post_start_service"))
	t.Equal(FirstPostStartContainer, WorkflowNameFromString("first_post_start_container"))
	t.Equal(FirstPostStartService, WorkflowNameFromString("first_post_start_service"))

	// Stop
	t.Equal(PreStopContainer, WorkflowNameFromString("pre_stop_container"))

	// Destroy
	t.Equal(PreDestroyContainer, WorkflowNameFromString("pre_destroy_container"))

	t.Equal(Undefined, WorkflowNameFromString("qwerty"))
}

func (t *WorkflowNameTestSuite) TestString() {
	// Start
	t.Equal("first_pre_start_container", FirstPreStartContainer.String())
	t.Equal("first_pre_start_service", FirstPreStartService.String())
	t.Equal("pre_start_container", PreStartContainer.String())
	t.Equal("pre_start_service", PreStartService.String())
	t.Equal("post_start_container", PostStartContainer.String())
	t.Equal("post_start_service", PostStartService.String())
	t.Equal("first_post_start_container", FirstPostStartContainer.String())
	t.Equal("first_post_start_service", FirstPostStartService.String())

	// Stop
	t.Equal("pre_stop_container", PreStopContainer.String())

	// Destroy
	t.Equal("pre_destroy_container", PreDestroyContainer.String())

	t.Equal("unknown", Undefined.String())
}
