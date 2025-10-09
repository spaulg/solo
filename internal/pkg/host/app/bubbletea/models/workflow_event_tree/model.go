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

	selectedWorkflowName  wms.WorkflowName
	selectedContainerName string

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
		for _, workflowName := range wms.WorkflowNames {
			if _, ok := t.workflowContainerMap[workflowName]; ok {
				t.selectedWorkflowName = workflowName
				break
			}
		}

		// Select first container of first workflow
		if t.selectedWorkflowName != wms.Undefined {
			containerNames := make([]string, 0)
			for containerName := range t.workflowContainerMap[t.selectedWorkflowName] {
				containerNames = append(containerNames, containerName)
			}

			if len(containerNames) > 0 {
				sort.Strings(containerNames)
				t.selectedContainerName = containerNames[0]
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
			t.selectedWorkflowName = wms.Undefined
			t.selectedContainerName = ""

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
			//containerName := t.workflowNames[t.selectedContainerIndex]
			//
			//workflowName := ""
			//if t.selectedWorkflowIndex >= 0 {
			//	workflowName = t.containerWorkflowMap[containerName][t.selectedWorkflowIndex]
			//}
			//
			//cmds = append(cmds, func() tea.Msg {
			//	return messages.WorkflowEventSelected{
			//		ExecutionEventName: t.executionEventName,
			//		ContainerName:      containerName,
			//		WorkflowName:       workflowName,
			//	}
			//})
		}
	}

	return t, tea.Batch(cmds...)
}

func (t Model) View() tea.View {
	treeView := tree.Root("")

	for _, workflowName := range wms.WorkflowNames {
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

		for _, containerName := range containerNames {
			if t.selectedContainerName == containerName {
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
	//	if t.selectedContainerIndex > 0 && t.selectedWorkflowIndex == -1 {
	//		t.selectedContainerIndex--
	//		t.selectedWorkflowIndex = len(t.containerWorkflowMap[t.workflowNames[t.selectedContainerIndex]]) - 1
	//
	//		return true
	//	} else if t.selectedWorkflowIndex >= 0 {
	//		t.selectedWorkflowIndex--
	//
	//		return true
	//	}

	return false
}

func (t *Model) selectNextRow() bool {
	//	containerName := t.workflowNames[t.selectedContainerIndex]
	//
	//	if t.selectedWorkflowIndex < len(t.containerWorkflowMap[containerName])-1 {
	//		t.selectedWorkflowIndex++
	//
	//		return true
	//	} else if t.selectedContainerIndex < len(t.workflowNames)-1 {
	//		t.selectedContainerIndex++
	//		t.selectedWorkflowIndex = -1
	//
	//		return true
	//	}

	return false
}
