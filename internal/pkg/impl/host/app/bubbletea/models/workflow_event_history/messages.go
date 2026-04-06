package workflow_event_history

import (
	"github.com/spaulg/solo/internal/pkg/impl/host/app/bubbletea/components/table"
)

type WorkflowHistoryDataLoaded struct {
	rows table.Rows
}
