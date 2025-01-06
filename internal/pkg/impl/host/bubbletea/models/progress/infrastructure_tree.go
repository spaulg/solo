package progress

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss/tree"

	progresscommon "github.com/spaulg/solo/internal/pkg/impl/common/container/progress"
	"github.com/spaulg/solo/internal/pkg/impl/host/bubbletea/messages"
)

type InfrastructureTreeModel struct {
	InfraTree *tree.Tree

	EntityTypeTree map[progresscommon.ProgressEntityTypeName]*tree.Tree
	EntityTree     map[progresscommon.ProgressEntityTypeName]map[string]*tree.Tree
	ContextTree    map[progresscommon.ProgressEntityTypeName]map[string]map[string]*tree.Tree
}

func NewInfrastructureTreeModel() tea.Model {
	return &InfrastructureTreeModel{
		InfraTree: tree.Root("Project"),

		EntityTypeTree: make(map[progresscommon.ProgressEntityTypeName]*tree.Tree),
		EntityTree:     make(map[progresscommon.ProgressEntityTypeName]map[string]*tree.Tree),
		ContextTree:    make(map[progresscommon.ProgressEntityTypeName]map[string]map[string]*tree.Tree),
	}
}

func (t *InfrastructureTreeModel) Init() tea.Cmd {
	return nil
}

func (t *InfrastructureTreeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {

	// todo: messages are not guaranteed to deliver in order
	//		 need to account for this by mapping the known status of an entity / action
	//		 use code from the status map to achieve this

	// todo: should I actually track the status inside the main code, then use a message
	//       to indicate an update has occurred, then sync the tree with the status map?
	//		 this would reduce the complexity of the messages needed to be raised

	case messages.ComposeProgressMsg:
		t.findOrCreateTreeBranch(m.EntityType, m.ProjectEntityName, m.ContextId)

	case messages.WorkflowStartedMsg:
		workflowName := m.WorkflowName.String()
		t.findOrCreateTreeBranch(progresscommon.Container, m.ContainerName, workflowName)

	case messages.WorkflowStepStartedMsg:
		workflowName := m.WorkflowName.String()
		t.findOrCreateTreeBranch(progresscommon.Container, m.ContainerName, workflowName)

	case messages.WorkflowStepOutputMsg:
		workflowName := m.WorkflowName.String()
		t.findOrCreateTreeBranch(progresscommon.Container, m.ContainerName, workflowName)

	case messages.WorkflowStepCompleteMsg:
		workflowName := m.WorkflowName.String()
		t.findOrCreateTreeBranch(progresscommon.Container, m.ContainerName, workflowName)

	case messages.WorkflowCompleteMsg:
		workflowName := m.WorkflowName.String()
		t.findOrCreateTreeBranch(progresscommon.Container, m.ContainerName, workflowName)
	}

	return t, nil
}

func (t *InfrastructureTreeModel) View() string {
	return t.InfraTree.String()
}

func (t *InfrastructureTreeModel) findOrCreateTreeBranch(
	entityType progresscommon.ProgressEntityTypeName,
	entityName string,
	contextId string,
) *tree.Tree {
	if _, exists := t.EntityTypeTree[entityType]; !exists {
		t.EntityTypeTree[entityType] = t.InfraTree.Child(entityType.String())
		t.EntityTree[entityType] = make(map[string]*tree.Tree)
		t.ContextTree[entityType] = make(map[string]map[string]*tree.Tree)
	}

	if _, exists := t.EntityTree[entityType][entityName]; !exists {
		t.EntityTree[entityType][entityName] = tree.New().Child(entityName)
		t.EntityTypeTree[entityType].Child(t.EntityTree[entityType][entityName])
		t.ContextTree[entityType][entityName] = make(map[string]*tree.Tree)
	}

	if _, exists := t.ContextTree[entityType][entityName][contextId]; !exists {
		t.ContextTree[entityType][entityName][contextId] = tree.New().Child(contextId)
		t.EntityTree[entityType][entityName].Child(t.ContextTree[entityType][entityName][contextId])
	}

	return t.ContextTree[entityType][entityName][contextId]
}
