package workflow_event_tree

import (
	"encoding/json"
	"os"
	"sort"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/tree"

	"github.com/spaulg/solo/internal/pkg/impl/common/domain/wms"

	messages2 "github.com/spaulg/solo/internal/pkg/impl/host/app/bubbletea/messages"
)

type Model struct {
	width  int
	height int

	containerNames []string
	containers     map[string][]string
}

type WorkflowMeta map[string][]string

type WorkflowEventLoaded struct {
	containerNames []string
	containers     map[string][]string
}

func NewModel() Model {
	return Model{
		containerNames: []string{},
		containers:     map[string][]string{},
	}
}

func (t Model) Init() tea.Cmd {
	return nil
}

func (t Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch m := msg.(type) {
	case messages2.ComponentSizeMsg:
		t.width, t.height = m.Width, m.Height

	case tea.KeyMsg:
		// todo: handle up, down and raising message to trigger update
		// 		 of associated output panel or use a callback

	case messages2.WorkflowEventSelected:
		cmds = append(cmds, func() tea.Msg {
			var containerNames []string
			containers := make(map[string][]string)

			// todo: list event directory contents
			workflowEventDirectory := ".solo/audit_logs/" + m.WorkflowEvent
			entries, err := os.ReadDir(workflowEventDirectory)
			if err != nil {
				// todo: handle error
				return nil
			}

			// todo: iterate for event names
			for _, entry := range entries {
				// Ignore non-directories
				if !entry.IsDir() {
					continue
				}

				// Filter for matching event names, ignoring anything else
				eventName := ""
				for _, workflowName := range wms.WorkflowNames {
					if entry.Name() == workflowName.String() {
						eventName = workflowName.String()
						break
					}
				}

				if eventName != "" {
					eventFile := ".solo/audit_logs/" + m.WorkflowEvent + "/" + entry.Name() + "/" + entry.Name() + ".meta.json"
					data, err := os.ReadFile(eventFile)
					if err != nil {
						// todo: handle error
						return nil
					}

					event := WorkflowMeta{}
					if err := json.Unmarshal(data, &event); err != nil {
						// todo: handle error
						return nil
					}

					// todo: for each event name directory found, read the associated
					//		 event {eventName}.meta.json file for container names
					for containerName := range event {
						if _, ok := containers[containerName]; !ok {
							containerNames = append(containerNames, containerName)
							containers[containerName] = make([]string, 0)
						}

						containers[containerName] = append(containers[containerName], eventName)
					}
				}
			}

			sort.Strings(containerNames)

			// todo: raise workflow events loaded - msg with container names & event names found
			return WorkflowEventLoaded{
				containerNames: containerNames,
				containers:     containers,
			}
		})

	case WorkflowEventLoaded:
		t.containers = m.containers
		t.containerNames = m.containerNames
	}

	return t, tea.Batch(cmds...)
}

func (t Model) View() tea.View {
	treeView := tree.Root("")

	for _, containerName := range t.containerNames {
		eventNames := t.containers[containerName]
		child := tree.New().Root(containerName)

		for _, event := range eventNames {
			child.Child(event)
		}

		treeView.Child(child)
	}

	return tea.NewView(lipgloss.NewStyle().Width(t.width).Height(t.height).Render(treeView.String()))
}
