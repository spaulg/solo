package progress

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type StatusTestSuite struct {
	suite.Suite
}

func TestStatusTestSuite(t *testing.T) {
	suite.Run(t, new(StatusTestSuite))
}

func (suite *StatusTestSuite) TestEmptyIDAndStatus() {
	progress := ComposeProgress{
		ID:     "",
		Status: "",
	}
	var expected *ComposeProgressEvent = nil

	event := progress.ToEvent("test_project")
	suite.Equal(expected, event)
}

func (suite *StatusTestSuite) TestIDWithLessThan2Parts() {
	progress := ComposeProgress{
		ID:     "singlepart",
		Status: "Running",
	}
	var expected *ComposeProgressEvent = nil

	event := progress.ToEvent("test_project")
	suite.Equal(expected, event)
}

func (suite *StatusTestSuite) TestBuiltIDWithLessThan2Parts() {
	progress := ComposeProgress{
		ID:     "singlepart",
		Status: "Built",
	}
	expected := &ComposeProgressEvent{
		Action: "Built",
		Type:   "Image",
		Entity: "singlepart",
	}

	event := progress.ToEvent("test_project")
	suite.Equal(expected, event)
}

func (suite *StatusTestSuite) TestIDWithMoreThan2Parts() {
	progress := ComposeProgress{
		ID:     "Container entity extra",
		Status: "Running",
	}
	expected := &ComposeProgressEvent{
		Action: "Running",
		Type:   "Container",
		Entity: "entity extra",
	}

	event := progress.ToEvent("test_project")
	suite.Equal(expected, event)
}

func (suite *StatusTestSuite) TestValidIDAndStatusWithHyphen() {
	progress := ComposeProgress{
		ID:     "Container test_project-entity",
		Status: "Running",
	}
	expected := &ComposeProgressEvent{
		Action: "Running",
		Type:   "Container",
		Entity: "entity",
	}

	event := progress.ToEvent("test_project")
	suite.Equal(expected, event)
}

func (suite *StatusTestSuite) TestValidIDAndStatusWithUnderscore() {
	progress := ComposeProgress{
		ID:     "Container test_project_entity",
		Status: "Running",
	}
	expected := &ComposeProgressEvent{
		Action: "Running",
		Type:   "Container",
		Entity: "entity",
	}

	event := progress.ToEvent("test_project")
	suite.Equal(expected, event)
}

func (suite *StatusTestSuite) TestValidIDAndStatusWithQuotes() {
	progress := ComposeProgress{
		ID:     "Container \"test_project-entity\"",
		Status: "Running",
	}
	expected := &ComposeProgressEvent{
		Action: "Running",
		Type:   "Container",
		Entity: "entity",
	}

	event := progress.ToEvent("test_project")
	suite.Equal(expected, event)
}
