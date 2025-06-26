package progress

import (
	"github.com/stretchr/testify/suite"

	progresscommon "github.com/spaulg/solo/internal/pkg/impl/common/container/progress"
)

type StatusTestSuite struct {
	suite.Suite
}

func (t *StatusTestSuite) TestEmptyIDAndStatus() {
	progress := ComposeProgress{
		ID:     "",
		Status: "",
	}

	event := progress.ToEvent("test_project")
	var expected *ComposeProgressEvent = nil

	t.Equal(expected, event)
}

func (t *StatusTestSuite) TestIDWithLessThan2Parts() {
	progress := ComposeProgress{
		ID:     "singlepart",
		Status: "Running",
	}

	event := progress.ToEvent("test_project")
	var expected *ComposeProgressEvent = nil

	t.Equal(expected, event)
}

func (t *StatusTestSuite) TestBuiltIDWithLessThan2Parts() {
	progress := ComposeProgress{
		ID:     "singlepart",
		Status: "Built",
	}

	event := progress.ToEvent("test_project")

	expected := &ComposeProgressEvent{
		ContextId:         event.ContextId,
		Action:            progresscommon.Build,
		EntityType:        progresscommon.Image,
		FullEntityName:    "singlepart",
		ProjectEntityName: "singlepart",
		Status:            progresscommon.Complete,
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
		ContextId:         event.ContextId,
		Action:            progresscommon.Create,
		EntityType:        progresscommon.Container,
		FullEntityName:    "entity extra",
		ProjectEntityName: "entity extra",
		Status:            progresscommon.InProgress,
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
		ContextId:         event.ContextId,
		Action:            progresscommon.Create,
		EntityType:        progresscommon.Container,
		FullEntityName:    "test_project-entity",
		ProjectEntityName: "entity",
		Status:            progresscommon.InProgress,
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
		ContextId:         event.ContextId,
		Action:            progresscommon.Create,
		EntityType:        progresscommon.Container,
		FullEntityName:    "test_project_entity",
		ProjectEntityName: "entity",
		Status:            progresscommon.InProgress,
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
		ContextId:         event.ContextId,
		Action:            progresscommon.Create,
		EntityType:        progresscommon.Container,
		FullEntityName:    "test_project-entity",
		ProjectEntityName: "entity",
		Status:            progresscommon.InProgress,
	}

	t.Equal(expected, event)
}
