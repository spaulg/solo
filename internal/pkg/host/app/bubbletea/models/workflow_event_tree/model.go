package workflow_event_tree

import (
	"maps"
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
	}
}

func NewModel(
	activeModel *layout.ActiveModel,
	containerStepMapRepository domain.ContainerStepMapRepository,
) Model {
	return Model{
		modelID:                    "workflow_event_tree",
		activeModel:                activeModel,
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
	var cmds []tea.Cmd

	switch m := msg.(type) {
	case messages.ComponentSizeMsg:
		t.width, t.height = m.Width, m.Height

	case messages.ExecutionEventHighlighted:
		cmds = append(cmds, func() tea.Msg {
			workflowContainerMap := make(map[wms.WorkflowName]map[string][]string)

			// todo: validate that the directory being loaded is for a workflow

			filepath := ".solo/audit_logs/" + m.ExecutionEvent
			for workflowNameString, containerStepMap := range t.containerStepMapRepository.Walk(
				filepath,
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
				executionEvent:       m.ExecutionEvent,
				workflowContainerMap: workflowContainerMap,
			}
		})

	case messages.ExecutionEventSelected:
		t.activeModel.SetActiveModel(t.modelID)

		// Select first workflow
		for workflowIndex, workflowName := range wms.WorkflowNames {
			if _, ok := t.workflowContainerMap[workflowName]; ok {
				t.selectedWorkflowIndex = workflowIndex
				break
			}
		}

		// Select first container of first workflow
		if t.selectedWorkflowIndex >= 0 {
			containerNames := make([]string, 0)
			for containerName := range t.workflowContainerMap[wms.WorkflowNames[t.selectedWorkflowIndex]] {
				containerNames = append(containerNames, containerName)
			}

			if len(containerNames) > 0 {
				sort.Strings(containerNames)
				t.selectedContainerIndex = 0
			}
		}

	case WorkflowEventLoaded:
		t.executionEventName = m.executionEvent
		t.workflowContainerMap = m.workflowContainerMap

	case tea.KeyMsg:
		// If we're not the active model, ignore key events
		if !t.activeModel.IsActive(t.modelID) {
			break
		}

		// If receiving an escape, deselect the execution event
		if key.Matches(m, t.keyMap.Escape) {
			t.selectedWorkflowIndex = -1
			t.selectedContainerIndex = -1

			cmds = append(cmds, func() tea.Msg {
				return messages.ExecutionEventDeselected{}
			})

			break
		}

		treeChanged := false

		switch {
		case key.Matches(m, t.keyMap.LineUp):
			treeChanged = t.selectPrevRow()
		case key.Matches(m, t.keyMap.LineDown):
			treeChanged = t.selectNextRow()
		}

		if treeChanged {
			workflowName := wms.WorkflowNames[t.selectedWorkflowIndex]
			containerNames := make([]string, 0)

			for containerName := range t.workflowContainerMap[wms.WorkflowNames[t.selectedWorkflowIndex]] {
				containerNames = append(containerNames, containerName)
			}

			sort.Strings(containerNames)

			cmds = append(cmds, func() tea.Msg {
				return messages.WorkflowEventSelected{
					ExecutionEventName: t.executionEventName,
					WorkflowName:       workflowName.String(),
					ContainerName:      containerNames[t.selectedContainerIndex],
				}
			})
		}
	}

	return t, tea.Batch(cmds...)
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
