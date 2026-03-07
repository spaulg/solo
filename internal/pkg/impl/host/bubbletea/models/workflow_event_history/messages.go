package workflow_event_history

import "github.com/spaulg/solo/internal/pkg/impl/host/bubbletea/components/table"

type WorkflowHistoryDataLoaded struct {
	rows table.Rows
}
