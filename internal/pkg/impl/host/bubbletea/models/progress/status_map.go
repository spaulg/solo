package progress

import (
	tea "github.com/charmbracelet/bubbletea"

	progresscommon "github.com/spaulg/solo/internal/pkg/impl/common/container/progress"
)

type ProgressStatus struct {
	EntityTypeOrder []progresscommon.ProgressEntityTypeName
	EntityOrder     map[progresscommon.ProgressEntityTypeName][]string
	ContextIdOrder  map[progresscommon.ProgressEntityTypeName]map[string][]string

	StatusMap map[progresscommon.ProgressEntityTypeName]map[string]map[string]tea.Model
}

func NewProgressStatus() *ProgressStatus {
	return &ProgressStatus{
		EntityTypeOrder: make([]progresscommon.ProgressEntityTypeName, 0),
		EntityOrder:     make(map[progresscommon.ProgressEntityTypeName][]string),
		ContextIdOrder:  make(map[progresscommon.ProgressEntityTypeName]map[string][]string),

		StatusMap: make(map[progresscommon.ProgressEntityTypeName]map[string]map[string]tea.Model),
	}
}

func (t *ProgressStatus) FindOrCreateStatusModel(
	entityType progresscommon.ProgressEntityTypeName,
	entityName string,
	contextId string,
	callback func() tea.Model,
) tea.Cmd {
	var cmd tea.Cmd

	if _, exists := t.StatusMap[entityType]; !exists {
		t.StatusMap[entityType] = make(map[string]map[string]tea.Model)
		t.EntityTypeOrder = append(t.EntityTypeOrder, entityType)
		t.EntityOrder[entityType] = make([]string, 0)
		t.ContextIdOrder[entityType] = make(map[string][]string)

	}

	if _, exists := t.StatusMap[entityType][entityName]; !exists {
		t.StatusMap[entityType][entityName] = make(map[string]tea.Model)
		t.EntityOrder[entityType] = append(t.EntityOrder[entityType], entityName)
		t.ContextIdOrder[entityType][entityName] = make([]string, 0)
	}

	if _, exists := t.StatusMap[entityType][entityName][contextId]; !exists {
		actionStatus := callback()

		t.StatusMap[entityType][entityName][contextId] = actionStatus
		t.ContextIdOrder[entityType][entityName] = append(t.ContextIdOrder[entityType][entityName], contextId)
		cmd = actionStatus.Init()
	}

	return cmd
}
