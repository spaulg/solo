package progress

import (
	"testing"

	"github.com/stretchr/testify/suite"

	progress2 "github.com/spaulg/solo/internal/pkg/shared/domain/container/progress"
)

func TestStatusTestSuite(t *testing.T) {
	suite.Run(t, new(StatusTestSuite))
}

type StatusTestSuite struct {
	suite.Suite
}

func (t *StatusTestSuite) TestEmptyIDAndStatus() {
	progress := ComposeProgress{
		ID:     "",
		Status: "",
	}

	event := progress.ToEvent("test_project")
	var expected *ComposeProgressEvent

	t.Equal(expected, event)
}

func (t *StatusTestSuite) TestIDWithLessThan2Parts() {
	progress := ComposeProgress{
		ID:     "singlepart",
		Status: "Running",
	}

	event := progress.ToEvent("test_project")
	var expected *ComposeProgressEvent

	t.Equal(expected, event)
}

func (t *StatusTestSuite) TestBuiltIDWithLessThan2Parts() {
	progress := ComposeProgress{
		ID:     "singlepart",
		Status: "Built",
	}

	event := progress.ToEvent("test_project")

	expected := &ComposeProgressEvent{
		ContextID:         event.ContextID,
		Action:            progress2.Build,
		EntityType:        progress2.Image,
		FullEntityName:    "singlepart",
		ProjectEntityName: "singlepart",
		Status:            progress2.Complete,
	}

	t.Equal(expected, event)
}

func (t *StatusTestSuite) TestIDWithMoreThan2Parts() {
	progress := ComposeProgress{
		ID:     "Container entity extra",
		Status: "Creating",
	}

	event := progress.ToEvent("test_project")

	expected := &ComposeProgressEvent{
		ContextID:         event.ContextID,
		Action:            progress2.Create,
		EntityType:        progress2.Container,
		FullEntityName:    "entity extra",
		ProjectEntityName: "entity extra",
		Status:            progress2.InProgress,
	}

	t.Equal(expected, event)
}

func (t *StatusTestSuite) TestValidIDAndStatusWithHyphen() {
	progress := ComposeProgress{
		ID:     "Container test_project-entity",
		Status: "Creating",
	}

	event := progress.ToEvent("test_project")

	expected := &ComposeProgressEvent{
		ContextID:         event.ContextID,
		Action:            progress2.Create,
		EntityType:        progress2.Container,
		FullEntityName:    "test_project-entity",
		ProjectEntityName: "entity",
		Status:            progress2.InProgress,
	}

	t.Equal(expected, event)
}

func (t *StatusTestSuite) TestValidIDAndStatusWithUnderscore() {
	progress := ComposeProgress{
		ID:     "Container test_project_entity",
		Status: "Creating",
	}

	event := progress.ToEvent("test_project")

	expected := &ComposeProgressEvent{
		ContextID:         event.ContextID,
		Action:            progress2.Create,
		EntityType:        progress2.Container,
		FullEntityName:    "test_project_entity",
		ProjectEntityName: "entity",
		Status:            progress2.InProgress,
	}

	t.Equal(expected, event)
}

func (t *StatusTestSuite) TestValidIDAndStatusWithQuotes() {
	progress := ComposeProgress{
		ID:     "Container \"test_project-entity\"",
		Status: "Creating",
	}

	event := progress.ToEvent("test_project")

	expected := &ComposeProgressEvent{
		ContextID:         event.ContextID,
		Action:            progress2.Create,
		EntityType:        progress2.Container,
		FullEntityName:    "test_project-entity",
		ProjectEntityName: "entity",
		Status:            progress2.InProgress,
	}

	t.Equal(expected, event)
}
