package workflow_event_output

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/suite"

	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/layout"
	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/messages"
	"github.com/spaulg/solo/internal/pkg/host/domain"
	"github.com/spaulg/solo/internal/pkg/host/infra/repository"
)

const executionEventName = "2026-11-16T19-48-40.655095032Z"
const workflowName = "first_pre_start_container"

func TestWorkflowEventOutputModelTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowEventOutputModelTestSuite))
}

type WorkflowEventOutputModelTestSuite struct {
	suite.Suite

	auditLogsPath string
	activeModel   *layout.ActiveModel
}

func (t *WorkflowEventOutputModelTestSuite) SetupTest() {
	t.auditLogsPath = t.T().TempDir()
	t.activeModel = layout.NewActiveModel()
	t.activeModel.SetActiveModel("workflow_event_tree")

	workflowPath := filepath.Join(t.auditLogsPath, executionEventName, workflowName)

	containerStepMapRepository := repository.NewJSONFileRepository[domain.ContainerStepMap]()
	t.Require().NoError(containerStepMapRepository.Save(
		filepath.Join(workflowPath, containerStepMapFile),
		domain.ContainerStepMap{
			"php-1":   {"STEP1", "STEP2", "STEP3"},
			"nginx-1": {"STEP4"},
		},
	))

	stepMetaRepository := repository.NewJSONFileRepository[*domain.WorkflowStepLogMeta]()
	for stepID, meta := range map[string]*domain.WorkflowStepLogMeta{
		"STEP1": {ExitCode: 0, Command: "/bin/sh", Arguments: []string{"-c", "echo hello"}, WorkingDirectory: "/tmp"},
		"STEP2": {ExitCode: 0, Command: "true"},
		"STEP3": {ExitCode: 2, Command: "/bin/false"},
		"STEP4": {ExitCode: 0, Command: "nginx", Arguments: []string{"-t"}},
	} {
		t.Require().NoError(stepMetaRepository.Save(filepath.Join(workflowPath, stepID+stepMetaFileSuffix), meta))
	}

	// STEP2 produced no output, so has no .out file
	t.Require().NoError(os.WriteFile(filepath.Join(workflowPath, "STEP1.out"), []byte("hello\r\nworld\n"), 0600))
	t.Require().NoError(os.WriteFile(filepath.Join(workflowPath, "STEP3.out"), []byte("something failed"), 0600))
	t.Require().NoError(os.WriteFile(filepath.Join(workflowPath, "STEP4.out"), []byte("nginx output\n"), 0600))
}

func (t *WorkflowEventOutputModelTestSuite) newModel() tea.Model {
	model := NewModel(
		t.activeModel,
		t.auditLogsPath,
		repository.NewJSONFileRepository[domain.ContainerStepMap](),
		repository.NewJSONFileRepository[*domain.WorkflowStepLogMeta](),
		repository.NewAppendFileStore(),
	)

	updated, _ := model.Update(messages.ComponentSizeMsg{Width: 120, Height: 30})

	return updated
}

func (t *WorkflowEventOutputModelTestSuite) selectContainer(model tea.Model, containerName string) tea.Model {
	model, cmd := model.Update(messages.WorkflowEventSelected{
		ExecutionEventName: executionEventName,
		WorkflowName:       workflowName,
		ContainerName:      containerName,
	})

	t.Contains(ansi.Strip(model.View().Content), "Loading")
	t.Require().NotNil(cmd)

	model, _ = model.Update(cmd())

	return model
}

func (t *WorkflowEventOutputModelTestSuite) TestRendersPlaceholderBeforeSelection() {
	view := ansi.Strip(t.newModel().View().Content)

	t.Contains(view, "Select a workflow container to view its output")
}

func (t *WorkflowEventOutputModelTestSuite) TestRendersCombinedOutputForContainerSteps() {
	view := ansi.Strip(t.selectContainer(t.newModel(), "php-1").View().Content)

	t.Contains(view, workflowName+" › php-1")
	t.Contains(view, `$ /bin/sh -c "echo hello" (cwd /tmp, exit 0)`)
	t.Contains(view, "hello")
	t.Contains(view, "world")
	t.Contains(view, "$ true (exit 0)")
	t.Contains(view, "(no output)")
	t.Contains(view, "$ /bin/false (exit 2)")
	t.Contains(view, "something failed")
	t.NotContains(view, "nginx output")

	t.Less(strings.Index(view, "echo hello"), strings.Index(view, "$ true"))
	t.Less(strings.Index(view, "$ true"), strings.Index(view, "/bin/false"))
}

func (t *WorkflowEventOutputModelTestSuite) TestIgnoresStaleLoadResult() {
	model := t.newModel()

	_, staleCmd := model.Update(messages.WorkflowEventSelected{
		ExecutionEventName: executionEventName,
		WorkflowName:       workflowName,
		ContainerName:      "php-1",
	})

	model = t.selectContainer(model, "nginx-1")
	model, _ = model.Update(staleCmd())

	view := ansi.Strip(model.View().Content)
	t.Contains(view, "nginx output")
	t.NotContains(view, "echo hello")
}

func (t *WorkflowEventOutputModelTestSuite) TestFocusAndUnfocus() {
	model := t.selectContainer(t.newModel(), "php-1")

	model, _ = model.Update(messages.WorkflowEventOutputFocused{})
	t.True(t.activeModel.IsActive("workflow_event_output"))
	t.Contains(ansi.Strip(model.View().Content), "esc to return")

	_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	t.Require().NotNil(cmd)
	t.IsType(messages.WorkflowEventOutputUnfocused{}, cmd())
}

func (t *WorkflowEventOutputModelTestSuite) TestClearsOnExecutionEventChange() {
	model := t.selectContainer(t.newModel(), "php-1")

	model, _ = model.Update(messages.ExecutionEventHighlighted{ExecutionEvent: "other"})

	t.Contains(ansi.Strip(model.View().Content), "Select a workflow container to view its output")
}
