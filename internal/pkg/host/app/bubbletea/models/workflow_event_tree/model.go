package workflow_event_tree

import (
	"maps"
	"path/filepath"
	"sort"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/tree"

	"github.com/spaulg/solo/internal/pkg/common/domain/wms"
	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/layout"
	"github.com/spaulg/solo/internal/pkg/host/domain"

	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/messages"
)

type Orientation int

type Model struct {
	activeModel *layout.ActiveModel

	modelID string
	width   int
	height  int

	auditLogsPath              string
	containerStepMapRepository domain.ContainerStepMapRepository

	executionEventName   string
	workflowContainerMap map[wms.WorkflowName]map[string][]string

	selectedWorkflowIndex  int
	selectedContainerIndex int

	keyMap KeyMap
	styles Styles
}

type WorkflowMeta map[string][]string

type WorkflowEventLoaded struct {
	executionEvent       string
	workflowContainerMap map[wms.WorkflowName]map[string][]string
}

func DefaultStyles() Styles {
	return Styles{
		Selected: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")),
		Node:     lipgloss.NewStyle(),
	}
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		LineUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		LineDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "view output"),
		),
	}
}

func NewModel(
	activeModel *layout.ActiveModel,
	auditLogsPath string,
	containerStepMapRepository domain.ContainerStepMapRepository,
) Model {
	return Model{
		modelID:                    "workflow_event_tree",
		activeModel:                activeModel,
		auditLogsPath:              auditLogsPath,
		containerStepMapRepository: containerStepMapRepository,
		workflowContainerMap:       make(map[wms.WorkflowName]map[string][]string),
		styles:                     DefaultStyles(),
		keyMap:                     DefaultKeyMap(),
		selectedWorkflowIndex:      -1,
		selectedContainerIndex:     -1,
	}
}

func (t Model) Init() tea.Cmd {
	return nil
}

func (t Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m := msg.(type) {
	case messages.ComponentSizeMsg:
		t.width, t.height = m.Width, m.Height

	case messages.ExecutionEventHighlighted:
		cmd = t.loadExecutionEvent(m.ExecutionEvent)

	case messages.ExecutionEventSelected:
		t.activeModel.SetActiveModel(t.modelID)
		cmd = t.selectFirstRow()

	case messages.WorkflowEventOutputUnfocused:
		t.activeModel.SetActiveModel(t.modelID)

	case WorkflowEventLoaded:
		t.executionEventName = m.executionEvent
		t.workflowContainerMap = m.workflowContainerMap

	case tea.KeyMsg:
		// If we're not the active model, ignore key events
		if t.activeModel.IsActive(t.modelID) {
			cmd = t.handleKey(m)
		}
	}

	return t, cmd
}

func (t Model) loadExecutionEvent(executionEvent string) tea.Cmd {
	return func() tea.Msg {
		workflowContainerMap := make(map[wms.WorkflowName]map[string][]string)

		// todo: validate that the directory being loaded is for a workflow

		executionEventPath := filepath.Join(t.auditLogsPath, executionEvent)
		for workflowNameString, containerStepMap := range t.containerStepMapRepository.Walk(
			executionEventPath,
			"container_step_map.meta.json",
		) {
			workflowName := wms.WorkflowNameFromString(workflowNameString)
			if workflowName == wms.Undefined {
				continue
			}

			if _, ok := workflowContainerMap[workflowName]; !ok {
				workflowContainerMap[workflowName] = make(map[string][]string)
			}

			maps.Copy(workflowContainerMap[workflowName], containerStepMap)
		}

		return WorkflowEventLoaded{
			executionEvent:       executionEvent,
			workflowContainerMap: workflowContainerMap,
		}
	}
}

// selectFirstRow selects the first container of the first workflow, if there is one.
func (t *Model) selectFirstRow() tea.Cmd {
	t.selectedWorkflowIndex = -1
	t.selectedContainerIndex = -1

	for workflowIndex, workflowName := range wms.WorkflowNames {
		if len(t.workflowContainerMap[workflowName]) > 0 {
			t.selectedWorkflowIndex = workflowIndex
			t.selectedContainerIndex = 0

			return t.workflowEventSelectedCmd()
		}
	}

	return nil
}

func (t *Model) handleKey(msg tea.KeyMsg) tea.Cmd {
	// If receiving an escape, deselect the execution event
	if key.Matches(msg, t.keyMap.Escape) {
		t.selectedWorkflowIndex = -1
		t.selectedContainerIndex = -1

		return func() tea.Msg {
			return messages.ExecutionEventDeselected{}
		}
	}

	// Nothing to navigate if the execution event has no workflows
	if t.selectedWorkflowIndex < 0 || t.selectedContainerIndex < 0 {
		return nil
	}

	treeChanged := false

	switch {
	case key.Matches(msg, t.keyMap.Enter):
		return func() tea.Msg {
			return messages.WorkflowEventOutputFocused{}
		}

	case key.Matches(msg, t.keyMap.LineUp):
		treeChanged = t.selectPrevRow()

	case key.Matches(msg, t.keyMap.LineDown):
		treeChanged = t.selectNextRow()
	}

	if treeChanged {
		return t.workflowEventSelectedCmd()
	}

	return nil
}

func (t Model) View() tea.View {
	treeView := tree.Root("")

	for workflowIndex, workflowName := range wms.WorkflowNames {
		if _, ok := t.workflowContainerMap[workflowName]; !ok {
			continue
		}

		child := tree.New()
		child = child.Root(workflowName.String())

		containerNames := make([]string, 0)
		for containerName := range t.workflowContainerMap[workflowName] {
			containerNames = append(containerNames, containerName)
		}

		sort.Strings(containerNames)

		for containerIndex, containerName := range containerNames {
			if t.selectedWorkflowIndex == workflowIndex && t.selectedContainerIndex == containerIndex {
				containerName = t.styles.Selected.Render(containerName)
			}

			child.Child(containerName)
		}

		treeView.Child(child)
	}

	return tea.NewView(
		lipgloss.NewStyle().
			Width(t.width).
			MaxWidth(t.width).
			Height(t.height).
			MaxHeight(t.height).
			Render(treeView.String()),
	)
}

func (t *Model) selectPrevRow() bool {
	if t.selectedWorkflowIndex < 0 {
		return false
	}

	if t.selectedContainerIndex > 0 {
		t.selectedContainerIndex--

		return true
	}

	// Advance previous workflow
	for workflowIndex := t.selectedWorkflowIndex - 1; workflowIndex >= 0; workflowIndex-- {
		workflowName := wms.WorkflowNames[workflowIndex]
		if _, ok := t.workflowContainerMap[workflowName]; ok {
			t.selectedWorkflowIndex = workflowIndex
			t.selectedContainerIndex = len(t.workflowContainerMap[workflowName]) - 1

			return true
		}
	}

	return false
}

func (t *Model) selectNextRow() bool {
	if t.selectedContainerIndex < len(t.workflowContainerMap[wms.WorkflowNames[t.selectedWorkflowIndex]])-1 {
		// Advance next container
		t.selectedContainerIndex++

		return true
	}

	// Advance next workflow
	for workflowIndex, workflowName := range wms.WorkflowNames[(t.selectedWorkflowIndex + 1):] {
		if _, ok := t.workflowContainerMap[workflowName]; ok {
			t.selectedWorkflowIndex += workflowIndex + 1
			t.selectedContainerIndex = 0

			return true
		}
	}

	return false
}

func (t Model) workflowEventSelectedCmd() tea.Cmd {
	workflowName := wms.WorkflowNames[t.selectedWorkflowIndex]
	containerNames := make([]string, 0)

	for containerName := range t.workflowContainerMap[workflowName] {
		containerNames = append(containerNames, containerName)
	}

	sort.Strings(containerNames)

	executionEventName := t.executionEventName
	containerName := containerNames[t.selectedContainerIndex]

	return func() tea.Msg {
		return messages.WorkflowEventSelected{
			ExecutionEventName: executionEventName,
			WorkflowName:       workflowName.String(),
			ContainerName:      containerName,
		}
	}
}
