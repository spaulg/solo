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
		Action:            Build,
		Type:              Image,
		FullEntityName:    "singlepart",
		ProjectEntityName: "singlepart",
		Status:            Complete,
	}

	event := progress.ToEvent("test_project")
	suite.Equal(expected, event)
}

func (suite *StatusTestSuite) TestIDWithMoreThan2Parts() {
	progress := ComposeProgress{
		ID:     "Container entity extra",
		Status: "Creating",
	}
	expected := &ComposeProgressEvent{
		Action:            Create,
		Type:              Container,
		FullEntityName:    "entity extra",
		ProjectEntityName: "entity extra",
		Status:            InProgress,
	}

	event := progress.ToEvent("test_project")
	suite.Equal(expected, event)
}

func (suite *StatusTestSuite) TestValidIDAndStatusWithHyphen() {
	progress := ComposeProgress{
		ID:     "Container test_project-entity",
		Status: "Creating",
	}
	expected := &ComposeProgressEvent{
		Action:            Create,
		Type:              Container,
		FullEntityName:    "test_project-entity",
		ProjectEntityName: "entity",
		Status:            InProgress,
	}

	event := progress.ToEvent("test_project")
	suite.Equal(expected, event)
}

func (suite *StatusTestSuite) TestValidIDAndStatusWithUnderscore() {
	progress := ComposeProgress{
		ID:     "Container test_project_entity",
		Status: "Creating",
	}
	expected := &ComposeProgressEvent{
		Action:            Create,
		Type:              Container,
		FullEntityName:    "test_project_entity",
		ProjectEntityName: "entity",
		Status:            InProgress,
	}

	event := progress.ToEvent("test_project")
	suite.Equal(expected, event)
}

func (suite *StatusTestSuite) TestValidIDAndStatusWithQuotes() {
	progress := ComposeProgress{
		ID:     "Container \"test_project-entity\"",
		Status: "Creating",
	}
	expected := &ComposeProgressEvent{
		Action:            Create,
		Type:              Container,
		FullEntityName:    "test_project-entity",
		ProjectEntityName: "entity",
		Status:            InProgress,
	}

	event := progress.ToEvent("test_project")
	suite.Equal(expected, event)
}
