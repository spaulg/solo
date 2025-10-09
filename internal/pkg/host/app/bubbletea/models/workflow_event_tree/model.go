package workflow_event_tree

import (
	"sort"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/tree"

	"github.com/spaulg/solo/internal/pkg/host/domain"

	"github.com/spaulg/solo/internal/pkg/host/app/bubbletea/messages"
)

const (
	containerFirstOrientation = iota
	workflowFirstOrientation
)

type Orientation int

type Model struct {
	width  int
	height int

	containerStepMapRepository domain.ContainerStepMapRepository

	executionEventName string

	orientation          Orientation
	workflowContainerMap map[string]map[string][]string

	keyMap KeyMap
	styles Styles
}

type WorkflowMeta map[string][]string

type WorkflowEventLoaded struct {
	executionEvent       string
	workflowContainerMap map[string]map[string][]string
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
	}
}

func NewModel(containerStepMapRepository domain.ContainerStepMapRepository) Model {
	return Model{
		containerStepMapRepository: containerStepMapRepository,
		workflowContainerMap:       make(map[string]map[string][]string),
		orientation:                containerFirstOrientation,
		styles:                     DefaultStyles(),
		keyMap:                     DefaultKeyMap(),
	}
}

func (t Model) Init() tea.Cmd {
	return nil
}

func (t Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch m := msg.(type) {
	case messages.ComponentSizeMsg:
		t.width, t.height = m.Width, m.Height

	case tea.KeyMsg:
		//treeChanged := false
		//
		//switch {
		//case key.Matches(m, t.keyMap.LineUp):
		//	treeChanged = t.selectPrevRow()
		//case key.Matches(m, t.keyMap.LineDown):
		//	treeChanged = t.selectNextRow()
		//}
		//
		//if treeChanged {
		//containerName := t.workflowNames[t.selectedContainerIndex]
		//
		//workflowName := ""
		//if t.selectedWorkflowIndex >= 0 {
		//	workflowName = t.containerWorkflowMap[containerName][t.selectedWorkflowIndex]
		//}

		//cmds = append(cmds, func() tea.Msg {
		//	return messages.WorkflowEventSelected{
		//		ExecutionEventName: t.executionEventName,
		//		ContainerName:      containerName,
		//		WorkflowName:       workflowName,
		//	}
		//})
		//}

	case messages.ExecutionEventSelected:
		cmds = append(cmds, func() tea.Msg {
			workflowContainerMap := make(map[string]map[string][]string)

			// todo: validate that the directory being loaded is for a workflow

			filepath := ".solo/audit_logs/" + m.ExecutionEvent
			for workflowName, containerStepMap := range t.containerStepMapRepository.Walk(filepath, "container_step_map.meta.json") {
				if t.orientation == containerFirstOrientation {
					// container name > workflow name > steps
					for containerName, steps := range containerStepMap {
						if _, ok := workflowContainerMap[containerName]; !ok {
							workflowContainerMap[containerName] = make(map[string][]string)
						}

						workflowContainerMap[containerName][workflowName] = steps
					}
				} else {
					// workflow name > container name > steps
					if _, ok := workflowContainerMap[workflowName]; !ok {
						workflowContainerMap[workflowName] = make(map[string][]string)
					}

					for containerName, steps := range containerStepMap {
						workflowContainerMap[workflowName][containerName] = steps
					}
				}
			}

			return WorkflowEventLoaded{
				executionEvent:       m.ExecutionEvent,
				workflowContainerMap: workflowContainerMap,
			}
		})

	case WorkflowEventLoaded:
		t.executionEventName = m.executionEvent
		t.workflowContainerMap = m.workflowContainerMap
	}

	return t, tea.Batch(cmds...)
}

func (t Model) View() tea.View {
	treeView := tree.Root("")

	level1Names := make([]string, 0)
	for level1Name := range t.workflowContainerMap {
		level1Names = append(level1Names, level1Name)
	}

	sort.Strings(level1Names)

	for _, level1Name := range level1Names {
		child := tree.New()

		rootLabel := level1Name
		//if t.selectedContainerIndex == workflowNameIndex && t.selectedWorkflowIndex == -1 {
		//	rootLabel = t.styles.Selected.Render(workflowName)
		//}

		child = child.Root(rootLabel)

		level2Names := make([]string, 0)
		for level2Name := range t.workflowContainerMap[level1Name] {
			level2Names = append(level2Names, level2Name)
		}

		sort.Strings(level2Names)

		for _, level2Name := range level2Names {
			childLabel := level2Name
			//	if t.selectedContainerIndex == workflowNameIndex && t.selectedWorkflowIndex == eventIndex {
			//		eventLabel = t.styles.Selected.Render(eventLabel)
			//	}

			child.Child(childLabel)
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

//func (t *Model) selectPrevRow() bool {
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
//
//	return false
//}
//
//func (t *Model) selectNextRow() bool {
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
//
//	return false
//}
